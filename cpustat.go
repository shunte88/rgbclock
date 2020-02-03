package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"io/ioutil"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type (
	cpuUsage struct {
		usage    float64
		total    float64
		idle     float64
		memtotal uint64
		memfree  uint64
	}

	// CPUStat simple CPU stats
	CPUStat struct {
		temperature   float64
		publish       cpuUsage
		last          cpuUsage
		csTimer       *time.Timer
		refresh       time.Duration
		cpustatstemp  draw.Image
		cpustatsusage draw.Image
		memstats      draw.Image
		mux           sync.Mutex
		face          font.Face
		fontHeight    float64
		color         color.Color
		lastTemp      float64
		lastUsage     float64
		lastMemFree   uint64
		update        bool
	}
)

// NewCPUStat initialise simple CPU stats widget
func NewCPUStat(f font.Face, x string) *CPUStat {

	cs := &CPUStat{
		temperature:   0,
		publish:       cpuUsage{0, 0, 0, 0, 0},
		last:          cpuUsage{0, 0, 0, 0, 0},
		refresh:       2 * time.Second, // htop freq.
		cpustatstemp:  image.NewRGBA(image.Rect(0, 0, 1, 1)),
		cpustatsusage: image.NewRGBA(image.Rect(0, 0, 1, 1)),
		memstats:      image.NewRGBA(image.Rect(0, 0, 1, 1)),
		lastTemp:      -999,
		lastUsage:     -999,
		update:        false,
	}

	cs.SetFace(f, x)
	cs.getCPUStats()
	return cs

}

// Stop deactivate stats scheduling
func (cs *CPUStat) Stop() {
	if cs.csTimer != nil {
		cs.csTimer.Stop()
	}
}

// CPUStatsTemp returns the CPU metrics glyph
func (cs *CPUStat) CPUStatsTemp() draw.Image {
	return cs.cpustatstemp
}

// CPUStatsUsage returns the CPU metrics glyph
func (cs *CPUStat) CPUStatsUsage() draw.Image {
	return cs.cpustatsusage
}

// MemStats returns the memory metrics glyph
func (cs *CPUStat) MemStats() draw.Image {
	return cs.memstats
}

// SetFace set font face and color
func (cs *CPUStat) SetFace(f font.Face, x string) {
	if cs.face != f {
		cs.face = f
		fmx := cs.face.Metrics()
		cs.fontHeight = float64((fmx.Height >> 6) + 2)
	}
	cs.color = parseHexColor(x)
}

func (cs *CPUStat) imageMetric(key, format string, size, hm int, metric float64) *image.RGBA {

	var ticon iconCache
	ticon, _ = cacheImage(key, ticon, 0.0, ``)
	dst := imaging.Resize(ticon.image, size, size, imaging.Lanczos)

	t := fmt.Sprintf(format, metric)
	adv := font.MeasureString(cs.face, t)
	canvas := image.NewRGBA(image.Rect(0, 0, (size + 1 + int(adv>>6)), hm))
	cb := canvas.Bounds()
	draw.Draw(canvas, cb, dst, cb.Min, draw.Src)

	d := &font.Drawer{
		Dst:  canvas,
		Src:  image.NewUniform(cs.color),
		Face: cs.face,
	}
	d.Dot = fixed.Point26_6{fixed.Int26_6((size + 1) * 64), fixed.Int26_6((cs.fontHeight - 4) * 64)}
	d.DrawString(t)

	return canvas

}

func (cs *CPUStat) drawStats() {

	if cs.lastTemp != cs.temperature ||
		cs.lastUsage != cs.publish.usage ||
		cs.lastMemFree != cs.publish.memfree {

		hc := int(1.05 * cs.fontHeight)
		size := int(hc - 2)

		if cs.lastTemp != cs.temperature {
			cs.cpustatstemp = cs.imageMetric(`cpu-temp`, "%.1f°C", size, hc, cs.temperature)
		}

		if cs.lastUsage != cs.publish.usage {
			cs.cpustatsusage = cs.imageMetric(`cpu-metrics`, "%.1f%%", size, hc, cs.publish.usage)
		}
		if cs.lastMemFree != cs.publish.memfree {
			cs.memstats = cs.imageMetric(`ram-metrics`, "%.1f%%", size, hc, 100.0*(float64(cs.publish.memfree)/float64(cs.publish.memtotal)))
		}

		cs.lastTemp = cs.temperature
		cs.lastUsage = cs.publish.usage
		cs.lastMemFree = cs.publish.memfree

	}
}

func (cs *CPUStat) getCPUStats() {

	cs.mux.Lock()

	cs.getMemStat()
	cs.getCPUTemp()
	idle, total := cs.getCPUSample()

	cs.publish.idle = cs.last.idle - float64(idle)
	cs.publish.total = cs.last.total - float64(total)

	if cs.publish.total != 0 {
		cs.publish.usage = 100.00 * (cs.publish.total - cs.publish.idle) / cs.publish.total
	} else {
		cs.publish.usage = 0.00
	}

	cs.last.idle = float64(idle)
	cs.last.total = float64(total)

	cs.mux.Unlock()

	go cs.drawStats()

	// and (re)init the repeat timer
	cs.csTimer = time.AfterFunc(cs.refresh, func() { cs.getCPUStats() })

}

func (cs *CPUStat) getCPUTemp() {

	path := `/sys/class/thermal/thermal_zone0/temp`
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("failed to read temperature from %q: %v", path, err)
		return
	}

	cpuTempStr := strings.TrimSpace(string(raw))
	cpuTempInt, err := strconv.Atoi(cpuTempStr)
	if err != nil {
		fmt.Printf("does not contain an integer: %v", path, err)
		return
	}

	cs.temperature = float64(cpuTempInt) / 1000.0
	return

}

func (cs *CPUStat) getCPUSample() (idle, total uint64) {

	raw, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}

	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if fields[0] == `cpu` {
			numFields := len(fields)
			for i := 1; i < numFields; i++ {
				val, err := strconv.ParseUint(fields[i], 10, 64)
				if err != nil {
					fmt.Println("Error: ", i, fields[i], err)
					continue
				}
				total += val // tally up all the numbers to get total ticks
				if i == 4 {  // idle is the 5th field in the cpu line
					idle = val
				}
			}
			return
		}
	}
	return
}

func (cs *CPUStat) getMemStat() {
	// Reference: man 5 proc, Documentation/filesystems/proc.txt in Linux source code
	raw, err := ioutil.ReadFile(`/proc/meminfo`)
	if err != nil {
		return
	}

	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		if strings.IndexRune(line, ':') < 0 {
			continue
		}
		fields := strings.Fields(strings.TrimSpace(strings.TrimRight(line, `kB`)))
		fields[0] = strings.TrimRight(fields[0], `:`)
		if v, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
			v = v * 1024
			switch fields[0] {
			case `MemTotal`:
				cs.publish.memtotal = v
			case `MemFree`:
				cs.publish.memfree = v
			}
		}

	}
}
