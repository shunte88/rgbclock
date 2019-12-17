package main

import (
	"fmt"
	"image"
	"image/draw"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"github.com/mellena1/mbta-v3-go/mbta"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// RGBPrediction holds realtime schedule information
type RGBPrediction struct {
	routeType     string
	stopName      string
	arrivalTime   time.Time
	departureTime time.Time
	direction     string
	routeID       string
	stopID        string
	color         string
	countdown     string
	stack         int
	changed       int // -1 invalid, 0 new, 1 update
	icon          iconCache
	slice         *image.RGBA
	canvas        *image.RGBA
}

// MBTA models transit predicted schedule
type MBTA struct {
	Display    bool
	mc         *mbta.Client
	days       []int
	active     bool
	update     chan bool
	mux        sync.Mutex
	offset     time.Duration
	route      []string
	stop       []string
	from       time.Time
	until      time.Time
	face       font.Face
	fontHeight float64
	prediction []RGBPrediction
}

// NewMBTAClient initiate an MBTA client
func NewMBTAClient(api, route, stop string, from, until time.Time, days []int, offset time.Duration) *MBTA {
	return &MBTA{
		mc:         mbta.NewClient(mbta.ClientConfig{APIKey: api}),
		active:     false,
		days:       days,
		offset:     offset,
		from:       from,
		until:      until,
		route:      cleanSplit(route),
		stop:       cleanSplit(stop),
		face:       basicfont.Face7x13,
		fontHeight: 13,
		Display:    false,
	}
}

func cleanSplit(s string) []string {
	ret := strings.Split(s, ",")
	for i := range ret {
		ret[i] = strings.TrimSpace(ret[i])
	}
	return ret
}

// Start updates
func (m *MBTA) Start() {
	if !m.active {
		m.update = sched(m.getPredicted, 30*time.Second)
		m.active = true
	}
}

// SetActiveHours implements predictions window
func (m *MBTA) SetActiveHours(from, until time.Time, days []int) {
	if m.from != from {
		m.from = from
	}
	if m.until != until {
		m.until = until
	}
	m.days = days

}

// Stop updates
func (m *MBTA) Stop() {
	if m.active {
		m.update <- true
		m.active = false
	}
}

func (m *MBTA) indexOf(pp RGBPrediction) int {
	for i, this := range m.prediction {
		if this.routeType == pp.routeType && this.stopName == pp.stopName && this.direction == pp.direction {
			return i
		}
	}
	return -1
}

func (m *MBTA) getPredicted() {

	day := int(time.Now().Weekday())
	check, _ := parseTime(time.Now().Format("03:04 PM"))
	if check.After(m.from) && check.Before(m.until) && intInSlice(m.days, day) {
		m.Display = true
	} else {
		m.Display = false
		return
	}

	pred, _, err := m.mc.Predictions.GetAllPredictions(
		&mbta.GetAllPredictionsRequestConfig{
			Sort:           mbta.PredictionsSortByArrivalTimeAscending,
			Include:        []mbta.PredictionInclude{"stop", "route"},
			FilterRouteIDs: m.route,
			FilterStopIDs:  m.stop,
		})

	if nil == err {

		// reset
		for i := range m.prediction {
			m.prediction[i].changed = -1
			m.prediction[i].countdown = ``
			m.prediction[i].stack = 0
		}
		testOff := time.Now().Add(m.offset)

		for _, p := range pred {

			if p.ScheduleRelationship != string(mbta.ScheduleRelationshipSkipped) {

				tt := p.ArrivalTime
				if "" == tt {
					tt = p.DepartureTime
				}
				testArr, err := time.Parse(time.RFC3339, tt)
				if nil == err && testArr.Before(testOff) && testArr.After(time.Now()) {

					if "" == p.DepartureTime {
						p.DepartureTime = p.ArrivalTime
					}
					testDep, err := time.Parse(time.RFC3339, p.DepartureTime)
					if nil != err {
						testDep = testArr
					}

					pn := p.Stop.Name
					if nil != p.Stop.PlatformName {
						pn += `.` + *p.Stop.PlatformName
					}

					pp := RGBPrediction{
						routeType:     p.Route.Description,
						stopName:      pn,
						arrivalTime:   testArr,
						departureTime: testDep,
						direction:     p.Route.DirectionNames[p.DirectionID],
						routeID:       m.fixRoute(p.Route.ID, p.Route.Description),
						stopID:        p.Stop.ID,
						changed:       0,
						countdown:     m.countdown(testArr),
						stack:         0,
						slice:         image.NewRGBA(image.Rect(0, 0, 1, 1)),
						canvas:        image.NewRGBA(image.Rect(0, 0, 1, 1)),
						color:         `#` + p.Route.Color,
					}
					idx := m.indexOf(pp)
					if -1 != idx {
						if m.prediction[idx].changed == -1 {
							m.prediction[idx].arrivalTime = pp.arrivalTime
							m.prediction[idx].departureTime = pp.departureTime
							m.prediction[idx].countdown = pp.countdown
							m.prediction[idx].changed = 1
						} else {
							m.prediction[idx].stack++
						}
					} else {
						m.prediction = append(m.prediction, pp)
					}
				} else {
					if nil != err {
						fmt.Println(err)
						fmt.Println(p)
					}
				}
			}
		}
		// update the graphic
		m.mux.Lock()
		for i, this := range m.prediction {
			if -1 != this.changed {

				m.prediction[i].changed = 1

				// init icon as needed
				if nil == m.prediction[i].icon.image {
					m.prediction[i].icon, _ = cacheImage(this.routeType, this.icon, 0.0, this.color)
				}

				m.prediction[i].canvas = image.NewRGBA(image.Rect(0, 0, 128, 18))

				dst := imaging.Resize(m.prediction[i].icon.image, 18, 18, imaging.Lanczos)
				draw.Draw(m.prediction[i].canvas, m.prediction[i].canvas.Bounds(), dst, image.Pt(0, 1), draw.Over)

				mx := m.face.Metrics()
				yy := 2 + int(float64(mx.Ascent>>6)-float64(mx.Descent>>6))
				point := fixed.Point26_6{fixed.Int26_6(19 * 64), fixed.Int26_6(yy * 64)}

				d := &font.Drawer{
					Dst:  m.prediction[i].canvas,
					Src:  image.NewUniform(parseHexColor(this.color)),
					Face: m.face,
					Dot:  point,
				}
				others := ``
				if this.stack > 0 {
					others = fmt.Sprintf(" +%d", this.stack)
				}
				d.DrawString(fmt.Sprintf("%v %v %v%v", this.routeID, this.direction, this.countdown, others))
				point = fixed.Point26_6{fixed.Int26_6(19 * 64), fixed.Int26_6((yy + 7) * 64)}
				d.Dot = point
				d.Src = image.NewUniform(image.White)
				d.DrawString(fmt.Sprintf("%v", this.stopName))
			}
		}
		m.mux.Unlock()
	} else {
		fmt.Println(err)
	}

}

func (m *MBTA) fixRoute(route, desc string) string {
	if `Rapid Transit` == desc {
		return route + ` Line`
	}
	return route
}

func (m *MBTA) countdown(t time.Time) string {
	ret := nextEvent(time.Now(), t)[3:]
	if strings.Contains(ret, `m`) {
		ret = strings.Split(ret, `m`)[0]
		if `1` == ret {
			ret += ` min`
		} else {
			ret += ` mins`
		}
	} else {
		ret = `at stop`
	}

	return ret
}

// SetFace defines font face
func (m *MBTA) SetFace(f font.Face) {
	if m.face != f {
		m.face = f
		fmx := m.face.Metrics()
		m.fontHeight = float64((fmx.Height >> 6) + 2)
	}
}

// Predictions enumerate active predictions
func (m *MBTA) Predictions() <-chan *image.RGBA {
	ch := make(chan *image.RGBA)
	go func() {
		for _, this := range m.prediction {
			if 1 == this.changed {
				ch <- this.canvas
			}
		}
		close(ch)
	}()
	return ch
}

func (m *MBTA) measureString(s string) float64 {
	d := &font.Drawer{
		Face: m.face,
	}
	adv := d.MeasureString(s)
	return float64(adv >> 6)
}

func (m *MBTA) parseHexColor(x string) (r, g, b, a int) {
	x = strings.TrimPrefix(x, "#")
	a = 255
	if len(x) == 3 {
		format := "%1x%1x%1x"
		fmt.Sscanf(x, format, &r, &g, &b)
		r |= r << 4
		g |= g << 4
		b |= b << 4
	}
	if len(x) == 6 {
		format := "%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b)
	}
	if len(x) == 8 {
		format := "%02x%02x%02x%02x"
		fmt.Sscanf(x, format, &r, &g, &b, &a)
	}
	return
}
