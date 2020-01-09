package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"

	"github.com/peterbourgon/diskv"
)

// CACache wraps the disk k/v connection and cache mechanism
type CACache struct {
	conn *diskv.Diskv
}

// InitImageCache initiates the diskv client
func InitImageCache(base string) *CACache {
	d := diskv.New(diskv.Options{
		BasePath:     base,
		CacheSizeMax: 10 * 1024 * 1024,
	})
	return &CACache{conn: d}
}

// Close the diskv client connection
func (car *CACache) Close() {
	car.conn.EraseAll() // we really want to timeout and clear - look at LRU instead???
}

// Get returns the response corresponding to key if present.
func (car *CACache) Get(key string) (resp []byte, ok bool) {
	ok = car.conn.Has(key)
	if ok {
		resp, _ = car.conn.Read(key)
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
