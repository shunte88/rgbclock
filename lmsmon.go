package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io/ioutil"
	"math"
	"net/http"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/disintegration/imaging"
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
	//PlaylistMode         string      `json:"playlist mode"`
	//PlaylistRepeat       int         `json:"playlist repeat"`
	//PlaylistShuffle      int         `json:"playlist shuffle"`
	//PlaylistCurIndex     interface{} `json:"playlist_cur_index"`
	PlaylistLoop []struct {
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
		PlaylistIndex  int         `json:"playlist index"`
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
	Bitrate     string
	Samplesize  float64
	Samplerate  float64
	Volume      int
	Year        string
}

// LMSServer limited to a single player for current usage
type LMSServer struct {
	id       int
	host     string
	port     int
	web      string
	url      string
	arturl   string
	coverart draw.Image
	Player   *LMSPlayer
	mux      sync.Mutex
	update   chan bool
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
		Artist:      NewInfoLabel(30, 1, d2, true, false),
		Album:       NewInfoLabel(30, 2, d1, true, false),
		Title:       NewInfoLabel(30, 1, d2, true, false),
		Albumartist: NewInfoLabel(30, 2, d1, true, false),
		Composer:    NewInfoLabel(30, 1, d2, true, false),
		Conductor:   NewInfoLabel(30, 1, d1, true, false),
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
		Volume:      0,
		Year:        `0`,
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
func NewLMSServer(host string, port int, player string) *LMSServer {
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
	ls.url = fmt.Sprintf("%s/jsonrpc.js", ls.web)
	ls.web += `/`
	ls.Player = NewLMSPlayer(player)
	return ls
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

				chka := ls.Player.Album.GetText()
				chkt := ls.Player.Artist.GetText()
				chkp := ls.Player.Title.GetText()

				ls.Player.Volume = s.MixerVolume

				if t, ok := s.Time.(float64); ok {
					ls.Player.setTime(t)
				}
				if d, ok := s.Duration.(float64); ok {
					ls.Player.setDuration(d)
				}

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
					//fmt.Printf("%v\n", s.PlaylistLoop[0].ArtworkURL)

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

				if chka != ls.Player.Album.GetText() ||
					chkt != ls.Player.Artist.GetText() ||
					chkp != ls.Player.Title.GetText() {
					ls.cacheImage()
				}
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

func (ls *LMSServer) cacheImage() {

	resp, err := http.Get(ls.arturl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	im, err := jpeg.Decode(resp.Body)
	if err == nil {
		ls.coverart = imaging.Resize(im, 500, 500, imaging.Lanczos)
	} else { // else try another tack
		im, err := png.Decode(resp.Body)
		if err == nil {
			ls.coverart = imaging.Resize(im, 500, 500, imaging.Lanczos)
		} else {
			resp.Body.Close()
			uri := fmt.Sprintf("http://%v:%v/music/0/cover_500x500_o", ls.host, ls.port)
			fmt.Println(uri)
			resp, err = http.Get(uri)
			if err != nil {
				panic(err)
			}
			defer resp.Body.Close()
			im, err = png.Decode(resp.Body)
			if err == nil {
				ls.coverart = imaging.Resize(im, 500, 500, imaging.Lanczos)
			} else {
				fmt.Println(err)
			}
		}
	}
}
