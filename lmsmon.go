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
	"math/rand"
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

type (
	// VUSetup - setup meter attributes
	VUSetup struct {
		color  string
		width  float64
		length float64 // percentile
		well   bool
	}
	// VULayout - visualization setup
	VULayout struct {
		meter     string
		mode      string
		base      string
		layout    string //  vertical/horizontal
		setup     VUSetup
		xpivot    [2]float64
		w         int
		h         int
		w2m       int
		h2m       int
		ptWidth   float64
		wMeter    float64
		rMeter    float64
		rWell     float64
		vu        draw.Image // think! this doubles memory foot print
		baseImage draw.Image
	}
	// LMSPlayer exposes several key attributes for the player and current track
	LMSPlayer struct {
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
	// SSES server details
	SSES struct {
		active   bool
		host     string
		port     int
		endpoint string
		url      string
		events   chan *SSEvent
	}
	// LMSServer limited to a single player for current usage
	LMSServer struct {
		id            int
		host          string
		port          int
		web           string
		url           string
		sses          SSES
		arturl        string
		coverart      draw.Image
		defaultart    draw.Image
		volume        draw.Image
		vulayout      VULayout
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
)

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
	// stop sses consumption
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

// LMSConfig setup
type LMSConfig struct {
	Host         string
	Port         int
	Player       string
	BaseFolder   string
	Meter        string
	MeterMode    string
	MeterLayout  string
	MeterBase    string
	NeedleColor  string
	NeedleWidth  float64
	NeedleLength float64 // percentile
	NeedleWell   bool
	SSESActive   bool
	SSESHost     string
	SSESPort     int
	SSESEndpoint string
}

// NewLMSServer initiates an LMS server instance
func NewLMSServer(lc LMSConfig) *LMSServer {
	ls := new(LMSServer)

	if `` == lc.Host {
		lc.Host = `localhost`
	}
	ls.id = 0
	ls.host = lc.Host
	ls.port = lc.Port
	ls.web = fmt.Sprintf("http://%s:%d", lc.Host, lc.Port)
	ls.arturl = fmt.Sprintf("%s/music/current/cover.jpg?player=%s", ls.web, lc.Player)
	ls.coverart = imaging.New(500, 500, color.NRGBA{0, 0, 0, 0})

	i := getIcon(`vinyl2`)
	i.scale = 500 / float64(i.width)
	ls.defaultart, _ = getImageIconWIP(i)

	ls.url = fmt.Sprintf("%s/jsonrpc.js", ls.web)
	ls.web += `/`
	ls.Player = NewLMSPlayer(lc.Player)
	ls.volume = image.NewRGBA(image.Rect(0, 0, 24, 16))
	ls.playmodifiers = image.NewRGBA(image.Rect(0, 0, 28, 16))
	ls.face = basicfont.Face7x13
	ls.fontHeight = 13
	ls.color = color.White
	ls.cacache = InitImageCache(lc.BaseFolder, true)

	ls.sses.active = lc.SSESActive
	ls.sses.host = lc.SSESHost
	ls.sses.port = lc.SSESPort
	if lc.SSESEndpoint[0] != '/' {
		lc.SSESEndpoint = `/` + lc.SSESEndpoint
	}
	ls.sses.endpoint = lc.SSESEndpoint
	ls.sses.url = fmt.Sprintf("%s:%d%s", lc.SSESHost, lc.SSESPort, lc.SSESEndpoint)

	ls.volviz = false
	ls.voltrig = time.NewTimer(2 * time.Second)
	ls.voltrig.Stop()
	ls.volinit = false

	ls.vulayout.vu = image.NewRGBA(image.Rect(0, 0, 1, 1)) // size as needed
	if `` != lc.Meter {
		ls.vulayout.meter = lc.Meter
		ls.vulayout.layout = lc.MeterLayout
		ls.vulayout.base = lc.MeterBase
		ls.vulayout.mode = lc.MeterMode
		ls.vulayout.setup.color = lc.NeedleColor
		ls.vulayout.setup.width = lc.NeedleWidth
		ls.vulayout.setup.length = lc.NeedleLength
		ls.vulayout.setup.well = lc.NeedleWell
		ls.initVUBase()
	}

	if ls.sses.active {
		ls.sseclient()
	}

	return ls

}

func (ls *LMSServer) sseclient() {
	ls.sses.events = make(chan *SSEvent)
	go ssenotifyacquire(ls.sses.url, ls.sses.events)
	go ls.consumeEvents()
}

func (ls *LMSServer) consumeEvents() {
	if ls.sses.active {

		wsa := lms.vulayout.baseImage.Bounds().Max.X
		hsa := lms.vulayout.baseImage.Bounds().Max.Y
		ssecanvas := image.NewRGBA(image.Rect(0, 0, wsa, hsa))
		govu := false

		wbin := wsa / (6 + (9 * 2))

		accum := [2]int32{-1, -1}
		dBfs := [2]int32{1000, 1000}
		dB := [2]int32{1000, 1000}
		linear := [2]int32{-1, -1}
		scaled := [2]int32{-1, -1}
		mindb := [2]int32{1000, 1000}
		maxdb := [2]int32{-1000, -1000}
		// fun with caps
		caps := [2][9]float64{{0, 0, 0, 0, 0, 0, 0, 0, 0}, {0, 0, 0, 0, 0, 0, 0, 0, 0}}

		// SA scaling
		multiSA := float64(hsa-1) / 31.00 // max input is 31 -2 to leave head-room

		for event := range ls.sses.events {

			// if we get an exception - review what happens here
			if !event.Active {
				ls.sses.active = false
				close(ls.sses.events)
				fmt.Println(`Inactive SSES, exit event stream`)
				return // and stop consuming...
			}

			var m Meter
			b, err := ioutil.ReadAll(event.Data)
			if err != nil {
				panic(err)
			}
			good := true
			if err := json.Unmarshal(b, &m); err != nil {
				fmt.Println(err) // observed incomplete JSON - timing thread safe pointers ???
				good = false
			}
			dirty := false
			dataset := m.Type

			if good {

				// much WIP experimental below
				for _, c := range m.Channels {
					i := 0
					if `R` == c.Name {
						i = 1
					}
					if c.DB < mindb[i] {
						mindb[i] = c.DB
						//fmt.Println(c.Name, `min`, mindb[i], `max`, maxdb[i])
					}
					if c.DB > maxdb[i] {
						maxdb[i] = c.DB
						//fmt.Println(c.Name, `min`, mindb[i], `max`, maxdb[i])
					}
					/*
						if accum[i] != c.Accumulated {
							dirty = true
							accum[i] = c.Accumulated
						}

						if linear[i] != c.Linear {
							dirty = true
							linear[i] = c.Linear
						}
					*/
					if dB[i] != c.DB {
						dirty = true
						dB[i] = c.DB
					}
					if dBfs[i] != c.DBfs {
						dirty = true
						dBfs[i] = c.DBfs
					}
					if scaled[i] != c.Scaled {
						dirty = true
						scaled[i] = c.Scaled
					}

				}
			}
			if dirty {

				if `VU` == dataset {
					govu = true
					if `vuPeak` != ls.vulayout.mode {
						ls.vuAnalog(accum, scaled, dB, dBfs, linear)
					} else {
						ls.vuPeak(dBfs)
						//ls.vuPeak(dB)
					}
				} else {

					if govu {
						draw.Draw(ssecanvas, ssecanvas.Bounds(), ls.vulayout.vu, image.ZP, draw.Src)
					} else {
						draw.Draw(ssecanvas, ssecanvas.Bounds(), ls.vulayout.baseImage, image.ZP, draw.Src)
					}

					oot := 0
					capll := color.RGBA{255, 0, 0, 192}
					for channel, c := range m.Channels {
						cll := getRandomColor(128)
						ofs := int(wbin/2) + 2 + (channel * (wsa / 2))
						for bin := 0; bin < 9; bin++ {
							test := 0.00
							if bin < int(c.NumFFT) {
								//cll = getRandomColor(64 + uint8((255-65)*(float64(c.FFT[bin])/31)))
								cll.A = 64 + uint8((255-65)*(float64(c.FFT[bin])/31))
								oot = int(multiSA * float64(c.FFT[bin]))
								test = float64(c.FFT[bin])
							} else {
								oot = int(multiSA / 4.00)
							}
							draw.Draw(ssecanvas, image.Rect(ofs, hsa, ofs+wbin-1, hsa-oot), &image.Uniform{cll}, image.ZP, draw.Src)
							if test >= caps[channel][bin] {
								caps[channel][bin] = test
							} else if caps[channel][bin] > 0 {
								caps[channel][bin]--
								if caps[channel][bin] < 0 {
									caps[channel][bin] = 0
								}
							}
							if caps[channel][bin] > 0 {
								coot := int(multiSA * caps[channel][bin])
								draw.Draw(ssecanvas, image.Rect(ofs, hsa-coot-1, ofs+wbin-1, hsa-coot), &image.Uniform{capll}, image.ZP, draw.Src)
							}
							ofs += (1 + wbin)
						}
					}
					ls.mux.Lock()
					draw.Draw(ls.vulayout.vu, ls.vulayout.vu.Bounds(), ssecanvas, image.ZP, draw.Src)
					ls.mux.Unlock()

				}
			}
		}
	}
}

func getRandomColor(alpha uint8) color.RGBA {
	rand.Seed(time.Now().UnixNano())
	r := uint8(rand.Intn(255))
	g := uint8(rand.Intn(255))
	b := uint8(rand.Intn(255))
	return color.RGBA{r, g, b, alpha}
}

// Close and clear the associated cache
func (ls *LMSServer) Close() {
	ls.cacache.Close()
}

// PlayerMAC sets player MAC - useful if current player changes
func (ls *LMSServer) PlayerMAC(player string) {
	ls.Player.MAC = player
}

// SSESAddress returns sses server address if active
func (ls *LMSServer) SSESAddress() string {
	r := ``
	if ls.sses.active {
		r = ls.sses.url
	}
	return r
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

// Stop the schedule updates and close event channel if active
func (ls *LMSServer) Stop() {
	ls.update <- true
	if ls.sses.active {
		close(ls.sses.events)
	}
	ls.Player.Stop()
}

func (ls *LMSServer) updatePlayer() {

	defer func() {
		if err := recover(); err != nil {
			ls.mux.Unlock() // ensure unlock
			fmt.Println(`exception recovered`, err)
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
					ls.Player.Album.SetText(title)
					ls.Player.Year = s.RemoteMeta.Year
					ls.Player.Genre = s.RemoteMeta.Genre
					ls.Player.coverid = s.RemoteMeta.Coverid
					ls.Player.Bitrate = s.RemoteMeta.Bitrate
					ls.Player.Bitty = fmt.Sprintf("• %v •", ls.Player.Bitrate)

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
					err = ls.cacheImage()
					checkFatal(err) // will be caught...
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

// VUActive meter is active
func (ls *LMSServer) VUActive() bool {
	return (ls.sses.active && `` != ls.vulayout.meter)
}

// VU returns the vu meter
func (ls *LMSServer) VU() draw.Image {
	return ls.vulayout.vu
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
		ls.drawBase(true)
		//im = ls.defaultart // ????????
		return im, err
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

	img = image.NewRGBA(image.Rect(0, 0, sw, sh))

	if ls.volviz {

		var iconMem = new(bytes.Buffer)
		var canvas = svg.New(iconMem)

		w, h := 80, 80
		canvas.Start(w, h)

		canvas.Group(`style="stroke:linen;stroke-width:0.2;fill-opacity:0.8;"`)
		canvas.Roundrect(5, 5, w-10, h-10, 6, 6, `style="fill:steelblue;"`)
		canvas.Roundrect(7, 7, w-14, h-14, 5, 5, `style="fill:none;"`)
		canvas.Gend()

		canvas.Group()

		var opacity [5]string
		switch {
		case ls.Player.Volume == 0:
			opacity = [5]string{"0.2", "0.2", "0.2", "0.2", "1"} // mute
		case ls.Player.Volume <= 10:
			opacity = [5]string{"0.2", "0.2", "0.2", "0.2", "0"} // lowest volume
		case ls.Player.Volume <= 20:
			opacity = [5]string{"0.9", "0.2", "0.2", "0.2", "0"} // one "bar"
		case ls.Player.Volume <= 40:
			opacity = [5]string{"0.9", "0.9", "0.2", "0.2", "0"} // two "bars"
		case ls.Player.Volume <= 80:
			opacity = [5]string{"0.9", "0.9", "0.9", "0.2", "0"} // three "bars"
		case ls.Player.Volume <= 100:
			opacity = [5]string{"0.9", "0.9", "0.9", "0.9", "0"} // highest volume - four "bars"
		}

		// cone
		canvas.Path(`m35.41333,44.18772l0,5.86667c0,1.76 -0.78222,3.12889 -2.15111,3.71556c-0.39111,0.19556 -0.58667,0.19556 -0.97778,0.19556c-0.78222,0 -1.56444,-0.39111 -2.15111,-0.97778l-8.8,-9.77778l-5.67111,0c-3.32444,0 -6.06222,-2.73778 -6.06222,-6.06222l0,-5.86667c0,-3.32444 2.73778,-6.06222 6.06222,-6.06222l5.67111,0l8.8,-9.77778c0.78222,-0.97778 1.95556,-1.17333 3.12889,-0.78222c1.36889,0.58667 2.15111,1.95556 2.15111,3.71556l0,6.06222l0,19.7511z`,
			`style="fill:navajowhite;fill-opacity:0.4;stroke:red;stroke-width:0.2"`)
		canvas.Path(`m31.69775,24.82771l0,-4.88889l-7.82222,8.8c-0.39111,0.39111 -0.97778,0.58667 -1.36889,0.58667l-6.84444,0c-1.17333,0 -2.15111,0.97778 -2.15111,2.15111l0,5.86667c0,1.17333 0.97778,2.15111 2.15111,2.15111l4.49778,0l0,-4.69333c0,-1.17333 0.78222,-1.95556 1.95556,-1.95556c0.97778,0 1.95556,0.78222 1.95556,1.95556l0,5.86667l7.43111,8.21333l0,-4.69333c0,-1.17333 0.97778,-1.95556 1.95556,-1.95556s1.95556,0.78222 1.95556,1.95556l0,5.86667c0,1.76 -0.78222,3.12889 -2.15111,3.71556c-0.39111,0.19556 -0.58667,0.19556 -0.97778,0.19556c-0.78222,0 -1.56444,-0.39111 -2.15111,-0.97778l-8.8,-9.77778l-5.67111,0c-3.32444,0 -6.06222,-2.73778 -6.06222,-6.06222l0,-5.86667c0,-3.32444 2.73778,-6.06222 6.06222,-6.06222l5.67111,0l8.8,-9.77778c0.78222,-0.97778 1.95556,-1.17333 3.12889,-0.78222c1.36889,0.58667 2.15111,1.95556 2.15111,3.71556l0,6.06222c0,0.97778 -0.97778,1.95556 -1.95556,1.95556s-1.76,-0.39111 -1.76,-1.56444l-0.00002,-0.00003z`,
			`style="fill:palegoldenrod;fill-opacity:0.5;"`)
		canvas.Path(`m33.6533,39.49776c-1.17333,0 -1.95556,-0.78222 -1.95556,-1.95556s0.78222,-1.95556 1.95556,-1.95556c0.78222,0 1.36889,-0.58667 1.36889,-1.36889s-0.58667,-1.36889 -1.36889,-1.36889c-1.17333,0 -1.95556,-0.78222 -1.95556,-1.95556s0.78222,-1.95556 1.95556,-1.95556c2.93333,0 5.28,2.34667 5.28,5.28s-2.54222,5.28 -5.28,5.28l0,0.00002z`,
			`style="fill:palegoldenrod;fill-opacity:0.9;"`)

		// volume bars
		for el := 0; el < 4; el++ {
			qx1 := 41 + (el * 6)
			qy1 := 29 - (el * 4)
			qy2 := 39 + (el * 4)
			qmx := float64(qx1) + float64(el+1)*3
			qmy := float64(qy1 + ((qy2 - qy1) / 2))
			canvas.Path(fmt.Sprintf("m%[1]d,%[2]dQ%[4]f,%[5]f %[1]d,%[3]d", qx1, qy1, qy2, qmx, qmy),
				fmt.Sprintf("style=\"fill:none;stroke:palegoldenrod;stroke-width:2.2;stroke-linecap:round;stroke-opacity:%s;\"", opacity[el]))
		}

		// mute
		canvas.Path(`m10.59556,44.13126zm54.80888,-1.67007l-8.00761,-8.00761l8.00616,-8.00761l-4.11587,-4.11878l-8.00761,8.00761l-8.00761,-8.00761l-4.11587,4.11878l8.00616,8.00761l-8.00761,8.00761l4.11878,4.11733l8.00616,-8.00761l8.00616,8.00761l4.11876,-4.11733z`,
			fmt.Sprintf("style=\"fill:crimson;fill-opacity:%s;\"", opacity[4]))

		canvas.Gend()

		canvas.Group()
		canvas.Line(12, h-12, w-12, h-12, `style="stroke-opacity:0.8;stroke-width:4.5;stroke:black;stroke-linecap:round"`)
		if ls.Player.Volume > 0 {
			x2 := int(float64((w-12)-12) * (float64(ls.Player.Volume) / 100.00))
			canvas.Line(12, h-12, 12+x2, h-12, `style="stroke-opacity:0.9;stroke-width:4.1;stroke:linen;stroke-linecap:round"`)
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
