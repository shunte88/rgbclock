package main

import (
	"time"

	"github.com/mmcdole/gofeed"
)

type freedres struct {
	feed *gofeed.Feed
	err  error
}

func news() {
	f := fetchFeeds()
	news := ``
	sep := ``
	for _, fl := range f {
		for _, fi := range fl.Items {
			if fi.PublishedParsed.After(lastNews) {
				//fmt.Printf("%v\n%v\n", fi.Title, fi.Description)
				news += sep + fi.Title
				sep = "\n"
			}
		}
	}
	lastNews = time.Now()
}

func fetchFeeds() []*gofeed.Feed {

	fc := make(chan freedres, len(feeds))

	for f := range feeds {
		feed := feeds[f].(map[interface{}]interface{})
		go fetchFeed(feed["link"].(string), fc)
	}

	var fs []*gofeed.Feed
	for i := 0; i < len(feeds); i++ {
		res := <-fc
		if res.err != nil {
			continue
		}
		fs = append(fs, res.feed)
	}

	return fs
}

func fetchFeed(uri string, fc chan freedres) {

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(uri)

	if err != nil {
		fc <- freedres{nil, err}
	}

	fc <- freedres{feed, nil}

}
