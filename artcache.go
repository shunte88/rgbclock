package main

import (
	"bytes"
	"container/list"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"

	"github.com/peterbourgon/diskv" // roll our own disk based LRU
)

type entry struct {
	key     string
	expires int64
}

// CACache wraps the k/v store cache insclusive LRU mechanism
type CACache struct {
	conn   *diskv.Diskv
	MaxAge int64
	mu     sync.Mutex
	lru    *list.List // Front is least-recent
	cache  map[string]*list.Element
	size   int64
}

// InitImageCache initiates the diskv client
func InitImageCache(base string, cleanup bool) *CACache {
	cac := &CACache{
		conn: diskv.New(diskv.Options{
			BasePath:     base,
			CacheSizeMax: 10 * 1024 * 1024,
		}),
		MaxAge: 2 * 60 * 60,
		lru:    list.New(),
		cache:  make(map[string]*list.Element),
	}
	if cleanup {
		_ = cac.cleanup()
	}
	return cac
}

func (car *CACache) isOlder(t time.Time, d time.Duration) bool {
	return time.Now().Sub(t) > d
}

func (car *CACache) cleanup() (err error) {

	tmpfiles, err := ioutil.ReadDir(car.conn.BasePath)
	if err != nil {
		return
	}

	for _, file := range tmpfiles {
		if file.Mode().IsRegular() {
			if car.isOlder(file.ModTime(), 24*time.Hour) {
				err = os.Remove(path.Join(car.conn.BasePath, file.Name()))
				if err != nil {
					return
				}
			}
		}
	}
	return
}

// Close the diskv client connection
func (car *CACache) Close() {
	fmt.Println(`not implemented`)
}

// Get returns the response corresponding to key if present.
func (car *CACache) Get(key string) (resp []byte, ok bool) {
	ok = car.conn.Has(key)
	if ok {
		resp, _ = car.conn.Read(key)

		le, okc := car.cache[key]
		if !okc {
			car.setLRU(key) // back-fill orphans
			return resp, ok
		}

		car.mu.Lock()
		if car.MaxAge > 0 && le.Value.(*entry).expires <= time.Now().Unix() {
			car.deleteElement(le)
			car.maybeDeleteOldest()
			car.mu.Unlock() // Avoiding defer overhead
			return nil, false
		}

		car.lru.MoveToBack(le)
		car.mu.Unlock() // Avoiding defer overhead
	}

	return resp, ok
}

// GetImage returns the image response corresponding to key if present.
func (car *CACache) GetImage(key string) (im image.Image, ok bool) {
	resp, ok := car.Get(key)
	if !ok {
		return nil, ok
	}

	im, _, err := image.Decode(bytes.NewReader(resp))
	if err != nil {
		fmt.Println(`caught`, err)
		return nil, false
	}
	return im, true

}

// Set saves a response to the cache as key.
func (car *CACache) Set(key string, resp []byte) {
	err := car.conn.Write(key, resp)
	if nil != err {
		fmt.Println(`set caught`, err)
	}
	car.setLRU(key)
}

func (car *CACache) setLRU(key string) {

	// LRU mech.
	expires := int64(0)
	if car.MaxAge > 0 {
		expires = time.Now().Unix() + car.MaxAge
	}

	car.mu.Lock()
	if le, ok := car.cache[key]; ok {
		car.lru.MoveToBack(le)
		e := le.Value.(*entry)
		e.expires = expires
	} else {
		e := &entry{key: key, expires: expires}
		car.cache[key] = car.lru.PushBack(e)
	}

	car.maybeDeleteOldest()
	car.mu.Unlock()

}

// SetImage saves an image to the keyyed cache.
func (car *CACache) SetImage(key string, im image.Image) {
	buff := new(bytes.Buffer)
	// note low quality - its an rgb panel so save the bytes
	err := jpeg.Encode(buff, im, &jpeg.Options{Quality: 70})
	if nil == err {
		car.Set(key, buff.Bytes())
	} else {
		fmt.Println(`caught`, err)
	}
}

func (car *CACache) maybeDeleteOldest() {

	if car.MaxAge > 0 {
		now := time.Now().Unix()
		for le := car.lru.Front(); le != nil && le.Value.(*entry).expires <= now; le = car.lru.Front() {
			car.deleteElement(le)
		}
	}

}

func (car *CACache) deleteElement(le *list.Element) {
	car.lru.Remove(le)
	e := le.Value.(*entry)
	delete(car.cache, e.key)
	err := car.conn.Erase(e.key)
	if nil != err {
		fmt.Println(`caught`, err)
	}
}
