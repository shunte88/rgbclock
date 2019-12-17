package main

import (
	"fmt"
	"image"
	"image/draw"
	"strings"
	"sync"
	"time"

	"github.com/fogleman/gg"
	"github.com/mmcdole/gofeed"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
)

// News configurator
type News struct {
	Feeds      []interface{}
	Detail     bool
	SeedTime   string
	atTime     time.Time
	endTime    time.Time
	Duration   time.Duration
	Repeat     time.Duration
	newsTimer  *time.Timer
	resetTimer *time.Timer
	lastNews   time.Time
	Velocity   int
	Width      int
	Limit      int
	news       string
	face       font.Face
	fontHeight float64
	mux        sync.Mutex
	bnd        image.Rectangle
	rct        image.Rectangle
	pt         image.Point
	marquee    chan bool
	slice      *image.RGBA
	canvas     *image.RGBA
	place      image.Rectangle
	display    bool
	active     bool
}

type freedres struct {
	feed *gofeed.Feed
	err  error
}

// InitNews instantiate news collector
func InitNews(nc News) *News {
	n := &News{
		Feeds:      nc.Feeds,
		Detail:     nc.Detail,
		SeedTime:   nc.SeedTime,
		Duration:   nc.Duration,
		Velocity:   nc.Velocity,
		Width:      nc.Width,
		Limit:      nc.Limit,
		atTime:     time.Now().Add(-1 * time.Hour),
		lastNews:   time.Now().Add(-24 * time.Hour),
		Repeat:     nc.Repeat,
		face:       basicfont.Face7x13,
		fontHeight: 13,
		active:     false,
	}

	if n.Velocity == 0 {
		n.Velocity = 6
	}
	if n.Width == 0 {
		n.Width = 126
	}

	n.nextTime()
	return n

}

// Stop deactivate news scheduling
func (n *News) Stop() {
	n.marquee <- true
	// deactivate schedule etc
	if n.resetTimer != nil {
		n.resetTimer.Stop()
	}
	if n.newsTimer != nil {
		n.newsTimer.Stop()
	}
}

func (n *News) nextTime() {

	// define next active window and init timers
	if n.atTime.Before(time.Now()) {
		n.atTime, _ = time.Parse(`2006-01-02T15:04-0700`,
			strings.Replace(time.Now().Format(`2006-01-02TXX:XX-0700`),
				`XX:XX`, n.SeedTime, -1))
	}

	for {
		if n.atTime.After(time.Now()) {
			break
		}
		n.atTime = n.atTime.Add(n.Repeat)
	}
	n.endTime = n.atTime.Add(n.Duration)

	waitDuration := time.Duration(n.atTime.Add(-1*time.Minute).UnixNano() - time.Now().UnixNano())
	n.newsTimer = time.AfterFunc(waitDuration, func() { n.getNews() })
	waitDuration = time.Duration(n.endTime.UnixNano() - time.Now().UnixNano())
	n.resetTimer = time.AfterFunc(waitDuration, func() { n.nextTime() })

	n.stopScroller()

}

// SetFace defines font face
func (n *News) SetFace(f font.Face) {
	if n.face != f {
		n.face = f
		fmx := n.face.Metrics()
		n.fontHeight = float64((fmx.Height >> 6) + 2)
	}
	if `` != n.news {
		n.paintCanvas()
	}
}

// Display init news scheduling
func (n *News) Display() bool {
	t := time.Now()
	return (t.After(n.atTime.Add(-1*time.Second)) &&
		t.Before(n.endTime) &&
		n.news != ``)
}

// getNews fetch news text
func (n *News) paintCanvas() {

	n.stopScroller()

	n.mux.Lock()
	const H = 2000
	const P = 2
	dc := gg.NewContext(n.Width, H)
	dc.Clear()
	dc.SetHexColor("#52bbc5cc")
	dc.SetFontFace(n.face)
	w, h := dc.MeasureMultilineString(strings.Join(dc.WordWrap(n.news, float64(n.Width-2)), "\n"), 1.2)
	dc.DrawStringWrapped(n.news, 1, 1, 0, 0, float64(n.Width-2), 1.2, gg.AlignLeft)

	nh := h + 5
	n.canvas = image.NewRGBA(image.Rect(0, 0, int(w), int(nh)))
	//n.canvas.
	n.bnd = n.canvas.Bounds()
	draw.Draw(n.canvas, n.bnd, dc.Image(), image.ZP, draw.Over)

	n.rct = n.bnd
	n.rct.Max.Y = n.Velocity
	n.slice = image.NewRGBA(n.rct)
	n.place = image.Rect(0, n.bnd.Max.Y-n.Velocity, n.bnd.Max.X, n.bnd.Max.Y)
	n.pt = image.Pt(0, n.Velocity)

	n.mux.Unlock()
	if int(nh) > n.Limit {
		n.initScroller()
	}
}

func (n *News) initScroller() {
	n.stopScroller()
	n.marquee = sched(n.scroller, 200*time.Millisecond)
	n.active = true
}

func (n *News) stopScroller() {
	if n.active {
		n.marquee <- true
	}
	n.active = false
}

// marquee scrolls text through canvas
func (n *News) scroller() {
	if n.canvas.Bounds().Max.Y > n.Limit {
		n.mux.Lock()
		draw.Draw(n.slice, n.rct, n.canvas, n.rct.Min, draw.Src)
		draw.Draw(n.canvas, n.bnd, n.canvas, n.bnd.Min.Add(n.pt), draw.Src)
		draw.Draw(n.canvas, n.place, n.slice, n.rct.Min, draw.Src)
		n.mux.Unlock()
	}

}

// getNews fetch news text
func (n *News) getNews() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	f := n.fetchFeeds()

	n.news = ``
	sep := ``
	for _, fl := range f {
		for _, fi := range fl.Items {
			if fi.PublishedParsed.After(n.lastNews) {
				if !strings.Contains(n.news, fi.Title) {
					n.news += sep + "• " + fi.Title
					if n.Detail {
						n.news += "\n" + fi.Description
					}
					sep = "\n"
				}
			}
		}
	}
	n.lastNews = time.Now()

	n.paintCanvas()

}

func (n *News) fetchFeeds() []*gofeed.Feed {

	fc := make(chan freedres, len(n.Feeds))

	for f := range n.Feeds {
		feed := n.Feeds[f].(map[interface{}]interface{})
		go n.fetchFeed(feed["link"].(string), fc)
	}

	var fs []*gofeed.Feed
	for i := 0; i < len(n.Feeds); i++ {
		res := <-fc
		if res.err != nil {
			continue
		}
		fs = append(fs, res.feed)
	}

	return fs
}

func (n *News) fetchFeed(uri string, fc chan freedres) {

	fp := gofeed.NewParser()
	feed, err := fp.ParseURL(uri)

	if err != nil {
		fc <- freedres{nil, err}
	}

	fc <- freedres{feed, nil}

}

// Image returns current canvas graphic-mode
func (n *News) Image() *image.RGBA {
	return n.canvas
}

func newsStop() {
	if nil != news {
		news.Stop()
	}
}
