package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"

	"github.com/gomodule/redigo/redis"
)

// CARedis wraps the redis connection and cache mechanism
type CARedis struct {
	conn redis.Conn
}

// InitImageCache initiates the redis client
func InitImageCache(host string, port int) (*CARedis, error) {
	srv := fmt.Sprintf("%s:%d", host, port)
	c, err := redis.Dial(`tcp`, srv)
	return &CARedis{conn: c}, err
}

// Close the redis client connection
func (car *CARedis) Close() {
	car.conn.Close()
}

// Get returns the response corresponding to key if present.
func (car *CARedis) Get(key string) (resp []byte, ok bool) {
	fmt.Println(`get`, key)
	item, err := redis.Bytes(car.conn.Do(`GET`, key))
	if err != nil {
		return nil, false
	}
	v, err := car.conn.Receive()
	fmt.Println(`get`, v, err)

	return item, true
}

// GetImage returns the image response corresponding to key if present.
func (car *CARedis) GetImage(key string) (im image.Image, ok bool) {
	fmt.Println(`getImage`, key)
	resp, ok := car.Get(key)
	if !ok {
		fmt.Println(key, `not found`)
		return nil, ok
	}
	fmt.Println(key, ok, len(resp))
	im, _, err := image.Decode(bytes.NewReader(resp))
	if err != nil {
		fmt.Println(err)
		return nil, false
	}
	return im, true

}

// Set saves a response to the cache as key.
func (car *CARedis) Set(key string, resp []byte) {
	fmt.Println(`set`, key, len(resp))
	car.conn.Do(`SET`, key, resp)
	car.conn.Flush()
	v, err := car.conn.Receive()
	fmt.Println(`set`, v, err)
}

// SetImage saves an image to the keyyed cache.
func (car *CARedis) SetImage(key string, im image.Image) {
	fmt.Println(`setImage`, key)
	buff := new(bytes.Buffer)
	err := jpeg.Encode(buff, im, &jpeg.Options{Quality: 95})
	if nil == err {
		car.Set(key, buff.Bytes())
	} else {
		fmt.Println(err)
	}
}
