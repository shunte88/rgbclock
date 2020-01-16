package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io"
	"io/ioutil"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/disintegration/imaging"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// LMSDetail provides current track detail and status
type LMSDetail struct {
	CanSeek              int         `json:"can_seek"`
	CurrentTitle         string      `json:"current_title,omitempty"`
	DigitalVolumeControl int         `json:"digital_volume_control"`
	Duration             interface{} `json:"duration"`
	MixerVolume          int         `json:"mixer volume"`
	Mode                 string      `json:"mode"`
	PlayerConnected      int         `json:"player_connected"`
	PlayerIP             string      `json:"player_ip"`
	PlayerName           string      `json:"player_name"`
	PlaylistRepeat       int         `json:"playlist repeat"`
	PlaylistShuffle      int         `json:"playlist shuffle"`
	PlaylistLoop         []struct {
		Album          string      `json:"album,omitempty"`
		Artist         string      `json:"artist,omitempty"`
		ArtworkURL     string      `json:"artwork_url"`
		AlbumartistIds string      `json:"albumartist_ids"`
		ArtworkTrackID string      `json:"artwork_track_id"`
		Albumartist    string      `json:"albumartist,omitempty"`
		Trackartist    string      `json:"trackartist,omitempty"`
		Conductor      string      `json:"conductor,omitempty"`
		Composer       string      `json:"composer,omitempty"`
		Bitrate        string      `json:"bitrate,omitempty"`
		Comment        string      `json:"comment,omitempty"`
		Compilation    string      `json:"compilation,omitempty"`
		Coverid        string      `json:"coverid"`
		Duration       interface{} `json:"duration"`
		Genre          string      `json:"genre,omitempty"`
		ID             interface{} `json:"id"`
		PlaylistIndex  interface{} `json:"playlist index"`
		Remote         string      `json:"remote,omitempty"`
		Samplesize     string      `json:"samplesize,omitempty"`
		Samplerate     string      `json:"samplerate,omitempty"`
		RemoteTitle    string      `json:"remote_title"`
		Title          string      `json:"title,omitempty"`
		Tracknum       string      `json:"tracknum,omitempty"`
		TrackartistIds string      `json:"trackartist_ids"`
		Type           string      `json:"type,omitempty"`
		URL            string      `json:"url,omitempty"`
		Year           string      `json:"year,omitempty"`
	} `json:"playlist_loop,omitempty"`
	PlaylistTimestamp float64 `json:"playlist_timestamp"`
	PlaylistTracks    int     `json:"playlist_tracks"`
	Power             int     `json:"power"`
	Rate              int     `json:"rate"`
	Remote            int     `json:"remote,omitempty"`
	RemoteMeta        struct {
		Artist      string      `json:"artist,omitempty"`
		Genre       string      `json:"genre,omitempty"`
		Bitrate     string      `json:"bitrate"`
		Coverid     string      `json:"coverid"`
		Duration    interface{} `json:"duration"`
		ID          interface{} `json:"id"`
		URL         string      `json:"url,omitempty"`
		Remote      int         `json:"remote"`
		RemoteTitle string      `json:"remote_title,omitempty"`
		Title       string      `json:"title,omitempty"`
		Year        string      `json:"year,omitempty"`
	} `json:"remoteMeta,omitempty"`
	SeqNo          int         `json:"seq_no"`
	Signalstrength int         `json:"signalstrength"`
	Time           interface{} `json:"time"`
}

//http://htmlpreview.github.io/?https://raw.githubusercontent.com/Logitech/slimserver/public/7.9/HTML/EN/html/docs/cli-api.html

// RPCError structure
type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// RPCRequest structure
type RPCRequest struct {
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
	ID     int         `json:"id"`
}

// RPCResponse structure
type RPCResponse struct {
	Result interface{} `json:"result,omitempty"`
	Error  *RPCError   `json:"error,omitempty"`
	ID     int         `json:"id"`
}

// NewRCPRequest "package" JSON RPC request
func NewRCPRequest(method string, params ...interface{}) *RPCRequest {
	request := &RPCRequest{
		Method: method,
		Params: Params(params...),
	}
	return request
}

// Params sanitize and initiate a parameter structure for JSON RPC
func Params(params ...interface{}) interface{} {

	var resultParams interface{}

	// if params was nil skip this and p stays nil
	if params != nil {
		switch len(params) {
		case 0: // no parameters, return nil
		case 1: // one param was provided, use it directly as is, or wrap primitive types in array
			if params[0] != nil {
				var typeOf reflect.Type

				// traverse until nil or not a pointer type
				for typeOf = reflect.TypeOf(params[0]); typeOf != nil && typeOf.Kind() == reflect.Ptr; typeOf = typeOf.Elem() {
				}

				if typeOf != nil {
					// now check if we can directly marshal the type or if it must be wrapped in an array
					switch typeOf.Kind() {
					// for these types we just do nothing, since value of p is already unwrapped from the array params
					case reflect.Struct:
						resultParams = params[0]
					case reflect.Array:
						resultParams = params[0]
					case reflect.Slice:
						resultParams = params[0]
					case reflect.Interface:
						resultParams = params[0]
					case reflect.Map:
						resultParams = params[0]
					default: // everything else must stay in an array (int, string, etc)
						resultParams = params
					}
				}
			} else {
				resultParams = params
			}
		default: // if more than one parameter was provided it should be treated as an array
			resultParams = params
		}
	}

	return resultParams
}

// LMSPlayer exposes several key attributes for the player and current track
type LMSPlayer struct {
	MAC         string
	Playername  string
	IP          string
	Mode        string
	Bitty       string // not Little Britain :)
	Artist      *InfoLabel
	Album       *InfoLabel
	Title       *InfoLabel
	Albumartist *InfoLabel
	Composer    *InfoLabel
	Conductor   *InfoLabel
	Compilation string
	Genre       string
	coverid     string
	time        float64
	TimeStr     string
	duration    float64
	DurStr      string
	remaining   float64
	RemStr      string
	Percent     float64
	remote      bool
	arturl      string
	coverart    draw.Image
	repeat      int
	shuffle     int
	Bitrate     string
	Samplesize  float64
	Samplerate  float64
	Volume      int
	Year        string
	lastVol     int
	lastRepeat  int
	lastShuffle int
}

// LMSServer limited to a single player for current usage
type LMSServer struct {
	id            int
	host          string
	port          int
	web           string
	url           string
	arturl        string
	coverart      draw.Image
	defaultart    draw.Image
	volume        draw.Image
	volviz        bool
	volinit       bool
	voltrig       *time.Timer
	playmodifiers draw.Image
	Player        *LMSPlayer
	mux           sync.Mutex
	face          font.Face
	fontHeight    float64
	color         color.Color
	cacache       *CACache
	update        chan bool
}

// NewLMSPlayer initiate an LMS player
func NewLMSPlayer(player string) *LMSPlayer {
	d1 := 120 * time.Millisecond
	d2 := 180 * time.Millisecond
	return &LMSPlayer{
		MAC:         player,
		Playername:  ``,
		IP:          ``,
		Mode:        ``,
		Bitty:       `CD`,
		Samplesize:  44.1,
		Samplerate:  16,
		Artist:      NewInfoLabel(34, 1, d2, true, false),
		Album:       NewInfoLabel(34, 2, d1, true, false),
		Title:       NewInfoLabel(34, 1, d2, true, false),
		Albumartist: NewInfoLabel(34, 2, d1, true, false),
		Composer:    NewInfoLabel(34, 1, d2, true, false),
		Conductor:   NewInfoLabel(34, 1, d1, true, false),
		Genre:       ``,
		coverid:     ``,
		time:        0.00,
		TimeStr:     `00:00`,
		duration:    0.00,
		DurStr:      `00:00`,
		remaining:   0.00,
		RemStr:      `00:00`,
		Percent:     0,
		remote:      false,
		repeat:      0,
		shuffle:     0,
		Volume:      0,
		Year:        `0`,
		lastVol:     -1,
		lastRepeat:  -1,
		lastShuffle: -1,
	}
}

func (p *LMSPlayer) setPercent() {
	if 0 != p.duration {
		p.Percent = 100 * p.time / p.duration
		p.remaining = p.duration - p.time
		p.RemStr = `-` + p.displayTime(p.remaining)
	} else {
		p.Percent = 0.00
		p.remaining = 0.00
		p.RemStr = `00:00`
	}
}

func (p *LMSPlayer) displayTime(t float64) string {
	hr := ``
	hours := math.Floor(t / 60 / 60)
	if hours > 0 {
		hr = fmt.Sprintf("%02.0f:", hours)
		t -= hours * 3600
	}
	return fmt.Sprintf("%s%02.0f:%02.0f", hr, math.Floor(t/60), math.Floor(math.Mod(t, 60)))
}

func (p *LMSPlayer) setDuration(d float64) {
	p.duration = d
	p.DurStr = p.displayTime(d)
	p.setPercent()
}

func (p *LMSPlayer) setTime(t float64) {
	p.time = t
	p.TimeStr = p.displayTime(t)
	p.setPercent()
}

// Stop marquee updates
func (p *LMSPlayer) Stop() {
	p.Artist.Stop()
	p.Album.Stop()
	p.Title.Stop()
	p.Albumartist.Stop()
	p.Composer.Stop()
	p.Conductor.Stop()
}

// Start marquee updates
func (p *LMSPlayer) Start() {
	p.Artist.Start()
	p.Album.Start()
	p.Title.Start()
	p.Albumartist.Start()
	p.Composer.Start()
	p.Conductor.Start()
}

// NewLMSServer initiates an LMS server instance
func NewLMSServer(host string, port int, player, base string) *LMSServer {
	ls := new(LMSServer)

	if `` == host {
		host = `localhost`
	}
	ls.id = 0
	ls.host = host
	ls.port = port
	ls.web = fmt.Sprintf("http://%s:%d", host, port)
	ls.arturl = fmt.Sprintf("%s/music/current/cover.jpg?player=%s", ls.web, player)
	ls.coverart = imaging.New(500, 500, color.NRGBA{0, 0, 0, 0})

	i := getIcon(`vinyl2`)
	i.scale = 500 / float64(i.width)
	ls.defaultart, _ = getImageIconWIP(i)

	ls.url = fmt.Sprintf("%s/jsonrpc.js", ls.web)
	ls.web += `/`
	ls.Player = NewLMSPlayer(player)
	ls.volume = image.NewRGBA(image.Rect(0, 0, 24, 16))
	ls.playmodifiers = image.NewRGBA(image.Rect(0, 0, 28, 16))
	ls.face = basicfont.Face7x13
	ls.fontHeight = 13
	ls.color = color.White
	ls.cacache = InitImageCache(base, true)

	ls.volviz = false
	ls.voltrig = time.NewTimer(2 * time.Second)
	ls.voltrig.Stop()
	ls.volinit = false

	return ls

}

// Close and clear the associated cache
func (ls *LMSServer) Close() {
	ls.cacache.Close()
}

// PlayerMAC sets player MAC - useful if current player changes
func (ls *LMSServer) PlayerMAC(player string) {
	ls.Player.MAC = player
}

func (ls *LMSServer) requestAndKey(params interface{}, key string) (interface{}, error) {
	ret, err := ls.request(ls.Player.MAC, params)
	if err != nil {
		return nil, err
	}
	if v, ok := ret.(map[string]interface{})[key]; ok {
		return v, nil
	}
	return nil, err // need a real error here

}

func (ls *LMSServer) request(player string, params interface{}) (interface{}, error) {

	if `` == player {
		player = `-`
	}

	data := &RPCRequest{
		Method: "slim.request",
		Params: Params(player, params),
		ID:     ls.id,
	}

	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", ls.url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	client := http.Client{Timeout: 10 * time.Second}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var resp RPCResponse
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Result, nil

}

// Start initiates the schedule update
func (ls *LMSServer) Start() {
	ls.update = sched(ls.updatePlayer, 400*time.Millisecond)
	ls.Player.Start()
}

// Stop the schedule update
func (ls *LMSServer) Stop() {
	ls.update <- true
	ls.Player.Stop()
}

func (ls *LMSServer) updatePlayer() {

	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()

	ls.mux.Lock()
	vs, err := ls.request(ls.Player.MAC, []string{"status", "-", "1", "tags:cgABbehldiqtyrSuoKLNJITC"})
	if nil != vs && err == nil {

		if v, ok := vs.(map[string]interface{}); ok {

			b, err := json.Marshal(v)
			if nil != err {
				panic(err)
			}
			s := LMSDetail{}
			err = json.Unmarshal(b, &s)
			if nil != err {
				panic(err)
			}
			ls.Player.Mode = s.Mode

			if `play` == s.Mode {

				ckcd := ls.Player.coverid
				ls.Player.Volume = s.MixerVolume

				if ls.Player.Volume != ls.Player.lastVol {
					ls.setVolume()
				}

				if t, ok := s.Time.(float64); ok {
					ls.Player.setTime(t)
				}
				if d, ok := s.Duration.(float64); ok {
					ls.Player.setDuration(d)
				}

				ls.Player.repeat = s.PlaylistRepeat
				ls.Player.shuffle = s.PlaylistShuffle

				if ls.Player.lastRepeat != ls.Player.repeat ||
					ls.Player.lastShuffle != ls.Player.shuffle {
					ls.setPlayModifiers()
				}

				ls.Player.lastVol = ls.Player.Volume
				ls.Player.lastRepeat = ls.Player.repeat
				ls.Player.lastShuffle = ls.Player.shuffle

				// remote
				ls.Player.remote = (1 == s.Remote)
				if ls.Player.remote {
					ls.Player.Artist.SetText(s.RemoteMeta.Artist)
					ls.Player.Albumartist.SetText(s.RemoteMeta.Artist)
					title := s.RemoteMeta.RemoteTitle
					if `` == title {
						title = s.RemoteMeta.Title
					}
					ls.Player.Title.SetText(title)
					ls.Player.Composer.SetText(``)
					ls.Player.Conductor.SetText(``)
					ls.Player.Year = s.RemoteMeta.Year
					ls.Player.Genre = s.RemoteMeta.Genre
					ls.Player.coverid = s.RemoteMeta.Coverid
					//fmt.Printf("%v\n", s.RemoteMeta.Coverid)
					//fmt.Printf("%v\n", s.RemoteMeta.URL)
				} else {
					artist := s.PlaylistLoop[0].Artist
					if artist == `` {
						artist = s.PlaylistLoop[0].Trackartist
					}
					ls.Player.Bitrate = s.PlaylistLoop[0].Bitrate
					f := s.PlaylistLoop[0].Samplesize
					ls.Player.Samplesize, err = strconv.ParseFloat(f, 64)
					if nil != err {
						ls.Player.Samplesize = 16
					}
					f = s.PlaylistLoop[0].Samplerate
					ls.Player.Samplerate, err = strconv.ParseFloat(f, 64)
					if nil != err {
						ls.Player.Samplerate = 44.1
					} else {
						ls.Player.Samplerate /= 1000
					}
					ls.Player.Artist.SetText(artist)
					ls.Player.Album.SetText(s.PlaylistLoop[0].Album)
					ls.Player.Composer.SetText(s.PlaylistLoop[0].Composer)
					ls.Player.Conductor.SetText(s.PlaylistLoop[0].Conductor)
					ls.Player.Title.SetText(s.PlaylistLoop[0].Title)
					ls.Player.Compilation = s.PlaylistLoop[0].Compilation
					if `1` == ls.Player.Compilation {
						ls.Player.Albumartist.SetText(`Various Artists`)
					} else {
						ls.Player.Albumartist.SetText(s.PlaylistLoop[0].Albumartist)
					}
					if `` == ls.Player.Albumartist.text {
						ls.Player.Albumartist.SetText(ls.Player.Artist.text)
					}
					ls.Player.Year = s.PlaylistLoop[0].Year
					ls.Player.Genre = s.PlaylistLoop[0].Genre
					ls.Player.coverid = s.PlaylistLoop[0].Coverid

					switch ls.Player.Samplesize {
					case 1:
						ls.Player.Bitty = fmt.Sprintf("• DSD%v •", math.Floor(ls.Player.Samplerate/44.1))
					default:
						ls.Player.Bitty = fmt.Sprintf("%vb • %vkHz", ls.Player.Samplesize, ls.Player.Samplerate)
					}

				}
				if "" == ls.Player.Year || "0" == ls.Player.Year {
					ls.Player.Year = "????"
				}

				if ckcd != ls.Player.coverid {
					//go ls.cacheImageBackground()
					err = ls.cacheImage()
					checkFatal(err)
				}
			} else {
				ls.volinit = false
			}

		} else {
			ls.Player.Mode = `unknown`
		}
	} else {
		ls.Player.Mode = `unknown`
	}
	ls.mux.Unlock()

}

// Coverart returns the cover image cache
func (ls *LMSServer) Coverart() draw.Image {
	return ls.coverart
}

// Volume returns the volume glyph
func (ls *LMSServer) Volume() draw.Image {
	return ls.volume
}

// PlayModifiers returns the playlist modifier glyph(s)
func (ls *LMSServer) PlayModifiers() draw.Image {
	return ls.playmodifiers
}

// SetFace set font face and color
func (ls *LMSServer) SetFace(f font.Face, x string) {
	if ls.face != f {
		ls.face = f
		fmx := ls.face.Metrics()
		ls.fontHeight = float64((fmx.Height >> 6) + 2)
	}
	ls.color = parseHexColor(x)
	ls.Player.Albumartist.SetFace(f, x)
	ls.Player.Album.SetFace(f, x)
	ls.Player.Title.SetFace(f, x)
	ls.Player.Artist.SetFace(f, x)
	ls.Player.Composer.SetFace(f, x)
	ls.Player.Conductor.SetFace(f, x)
}

// SetMaxLen set scroll limits
func (ls *LMSServer) SetMaxLen(m int) {
	ls.Player.Albumartist.SetMaxlen(m)
	ls.Player.Album.SetMaxlen(m)
	ls.Player.Title.SetMaxlen(m)
	ls.Player.Artist.SetMaxlen(m)
	ls.Player.Composer.SetMaxlen(m)
	ls.Player.Conductor.SetMaxlen(m)
}

func (ls *LMSServer) setVolume() {

	var ticon iconCache
	if 0 == ls.Player.Volume {
		ticon, _ = cacheImage(`volume-mute`, ticon, 0.0, ``)
	} else {
		for i := 0; i < 5; i++ {
			if ls.Player.Volume <= ((i + 1) * 20) {
				ticon, _ = cacheImage(fmt.Sprintf("volume-%d", i), ticon, 0.0, ``)
				break
			}
		}
	}
	wc := int(3.5 * ls.fontHeight)
	hc := int(1.5*ls.fontHeight) + 2
	canvas := image.NewRGBA(image.Rect(0, 0, wc, hc))
	size := hc - 5
	dst := imaging.Resize(ticon.image, size, size, imaging.Lanczos)
	cb := canvas.Bounds()
	cb.Min.X = int(float64(cb.Max.X) * .333)

	draw.Draw(canvas, cb, dst, image.ZP, draw.Src)

	if 0 != ls.Player.Volume {
		t := fmt.Sprintf("%d", ls.Player.Volume)
		d := &font.Drawer{
			Dst:  canvas,
			Src:  image.NewUniform(ls.color),
			Face: ls.face,
		}
		adv := d.MeasureString(t)
		fw := float64(cb.Min.X-1) - float64(adv>>6) // right justify
		d.Dot = fixed.Point26_6{fixed.Int26_6(fw * 64), fixed.Int26_6(9 * 64)}
		d.DrawString(t)
	}

	draw.Draw(ls.volume, ls.volume.Bounds(), canvas, image.ZP, draw.Src)

	if ls.volinit {
		ls.voltrig.Reset(2 * time.Second)
		ls.volviz = true
		go func() {
			select {
			case <-ls.voltrig.C:
				ls.volviz = false
			}
		}()
	}
	ls.volinit = true

}

func (ls *LMSServer) setPlayModifiers() {

	wc := int(3 * ls.fontHeight)
	hc := int(1.5*ls.fontHeight) + 2
	canvas := image.NewRGBA(image.Rect(0, 0, wc, hc))
	size := hc - 4

	var ticon iconCache
	cb := canvas.Bounds()
	cb.Min.X = 2
	if 0 != ls.Player.repeat {
		ticon, _ = cacheImage(fmt.Sprintf("repeat-%d", ls.Player.repeat), ticon, 0.0, ``)
		dst := imaging.Resize(ticon.image, size, size, imaging.Lanczos)
		draw.Draw(canvas, cb, dst, image.ZP, draw.Src)
	}
	cb.Min.X += size
	if 0 != ls.Player.shuffle {
		ticon, _ = cacheImage(fmt.Sprintf("shuffle-%d", ls.Player.shuffle), ticon, 0.0, ``)
		dst := imaging.Resize(ticon.image, size, size, imaging.Lanczos)
		draw.Draw(canvas, cb, dst, image.ZP, draw.Src)
	}

	draw.Draw(ls.playmodifiers, ls.playmodifiers.Bounds(), canvas, image.ZP, draw.Src)
}

func (ls *LMSServer) getImage(r io.Reader) (image.Image, error) {

	// need to make this buffered and cancel-able!
	im, err := imaging.Decode(r, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	im = imaging.Resize(im, 500, 500, imaging.Lanczos)
	return im, nil

}

func (ls *LMSServer) drawBase(defart bool) {
	if defart {
		draw.Draw(ls.coverart, ls.coverart.Bounds(), ls.defaultart, image.ZP, draw.Src)
	} else {
		draw.Draw(ls.coverart, ls.coverart.Bounds(), &image.Uniform{color.Black}, image.ZP, draw.Src)
	}
}

func (ls *LMSServer) cacheImageBackground() {
	err := ls.cacheImage()
	checkFatal(err)
}

func (ls *LMSServer) cacheImage() error {

	// check if we have the cover cached
	im, ok := ls.cacache.GetImage(ls.Player.coverid)
	if ok {

		ls.drawBase(false)
		draw.Draw(ls.coverart, ls.coverart.Bounds(), im, image.ZP, draw.Src)
		return nil

	}

	resp, err := http.Get(ls.arturl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// chunked or large images blank the coverart
	//slow load on init but we cache the thumbnail
	cl, _ := strconv.ParseInt(resp.Header.Get(`content-length`), 10, 64)
	if cl == -1 || cl > 1000000 {
		ls.drawBase(true)
	}

	im, err = ls.getImage(resp.Body)
	if err != nil {
		return err
	}

	if im == nil {

		resp.Body.Close()
		resp, err = http.Get(fmt.Sprintf("http://%v:%v/music/0/cover_500x500_o", ls.host, ls.port))
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		im, err = ls.getImage(resp.Body)
		if err != nil {
			return err
		}

	}

	if im != nil {
		draw.Draw(ls.coverart, ls.coverart.Bounds(), im, image.ZP, draw.Src)
		ls.cacache.SetImage(ls.Player.coverid, im)
	}

	return nil

}

// VolumePopup - visualize volume change - a la Ubuntu desktop ;)
func (ls *LMSServer) VolumePopup(sw, sh int) (img draw.Image) {

	w, h := 80, 80

	img = image.NewRGBA(image.Rect(0, 0, sw, sh))

	if ls.volviz {
		var iconMem = new(bytes.Buffer)

		var canvas = svg.New(iconMem)
		canvas.Start(w, h)

		canvas.Group(`style="stroke:linen;stroke-width:0.2;fill-opacity:0.8;"`)
		canvas.Roundrect(5, 5, 70, 70, 6, 6, `style="fill:steelblue;"`)
		canvas.Roundrect(7, 7, 66, 66, 5, 5, `style="fill:none;"`)
		canvas.Gend()

		canvas.Group()

		var opacity [5]string
		switch {
		case ls.Player.Volume == 0:
			opacity = [5]string{"0.2", "0.2", "0.2", "0.2", "1"}
		case ls.Player.Volume <= 10:
			opacity = [5]string{"0.2", "0.2", "0.2", "0.2", "0"}
		case ls.Player.Volume <= 20:
			opacity = [5]string{"1", "0.2", "0.2", "0.2", "0"}
		case ls.Player.Volume <= 40:
			opacity = [5]string{"1", "1", "0.2", "0.2", "0"}
		case ls.Player.Volume <= 80:
			opacity = [5]string{"1", "1", "1", "0.2", "0"}
		case ls.Player.Volume <= 100:
			opacity = [5]string{"1", "1", "1", "1", "0"}
		}

		// cone
		canvas.Path(`m9.44252,25.5513l0,17.8031l10.13487,0l15.17187,12.45947l0.00902,-42.72203l-15.17412,12.45947l-10.14163,0l-0.00001,-0.00001z`,
			`style="fill:palegoldenrod;"`)

		// volume bars
		for el := 0; el < 4; el++ {
			canvas.Line(42+(el*8), 29-(el*4), 42+(el*8), 39+(el*4),
				fmt.Sprintf("style=\"fill:none;stroke:palegoldenrod;stroke-width:3.5;stroke-linecap:round;stroke-opacity:%s;\"", opacity[el]))
		}

		// mute
		canvas.Path(`m10.59556,46.1477zm54.80888,-1.67007l-8.00761,-8.00761l8.00616,-8.00761l-4.11587,-4.11878l-8.00761,8.00761l-8.00761,-8.00761l-4.11587,4.11878l8.00616,8.00761l-8.00761,8.00761l4.11878,4.11733l8.00616,-8.00761l8.00616,8.00761l4.11876,-4.11733z`,
			fmt.Sprintf("style=\"fill:crimson;fill-opacity:%s;\"", opacity[4]))

		canvas.Gend()

		canvas.Group()
		canvas.Line(12, 68, 68, 68, `style="stroke-opacity:0.8;stroke-width:4.5;stroke:black;stroke-linecap:round"`)
		if ls.Player.Volume > 0 {
			x2 := int(float64(68-12) * (float64(ls.Player.Volume) / 100.00))
			canvas.Line(12, 68, 12+x2, 68, `style="stroke-opacity:0.9;stroke-width:4.1;stroke:linen;stroke-linecap:round"`)
		}
		canvas.Gend()

		canvas.End()

		//fmt.Println(iconMem.String())
		iconI, err := oksvg.ReadIconStream(iconMem)
		if err != nil {
			return img
		}

		gv := rasterx.NewScannerGV(w, h, img, img.Bounds())
		r := rasterx.NewDasher(w, h, gv)
		iconI.SetTarget(0, 0, float64(sw), float64(sh))
		iconI.Draw(r, 1.0)
	}

	return img

}
