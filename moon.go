package main

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	"math"
	"time"

	svg "github.com/ajstarks/svgo"
	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

// Moon calculations
type Moon struct {
	phase     float64
	illum     float64
	age       float64
	dist      float64
	angdia    float64
	sundist   float64
	sunangdia float64
	pdata     float64
	quarters  [8]float64
	timespace float64
	azimuth   float64
	altitude  float64
	distance  float64
}

const millyToNano = 1000000
const dayMs = 1000 * 60 * 60 * 24
const J1970 = 2440588
const J2000 = 2451545

func timeToUnixMillis(date time.Time) int64   { return int64(float64(date.UnixNano()) / millyToNano) }
func unixMillisToTime(date float64) time.Time { return time.Unix(0, int64(date*millyToNano)) }
func toJulian(date time.Time) float64         { return float64(timeToUnixMillis(date))/dayMs - 0.5 + J1970 }
func fromJulian(j float64) time.Time          { return unixMillisToTime((j + 0.5 - J1970) * dayMs) }
func toDays(date time.Time) float64           { return toJulian(date) - J2000 }

// general calculations for position
const rad = math.Pi / 180
const e = rad * 23.4397              // obliquity of the Earth
const synmonth float64 = 29.53058868 // Synodic month (new Moon to new Moon)

// NewLuna creates an instance of the Luna impl. develops math - DOEST NOT INCLUDE OBSERVATION POINT!
func NewLuna(t time.Time, lat, lng float64) (moonP *Moon) {

	moonP = new(Moon)

	// Astronomical constants
	var epoch float64 = 2444238.5 // 1989 January 0.0

	//Constants defining the Sun's apparent orbit
	var elonge float64 = 278.833540  // Ecliptic longitude of the Sun at epoch 1980.0
	var elongp float64 = 282.596403  // Ecliptic longitude of the Sun at perigee
	var eccent float64 = 0.016718    // Eccentricity of Earth's orbit
	var sunsmax float64 = 1.495985e8 // Sun's angular size, degrees, at semi-major axis distance
	var sunangsiz float64 = 0.533128

	// Elements of the Moon's orbit, epoch 1980.0
	var mmlong float64 = 64.975464   // Moon's mean longitude at the epoch
	var mmlongp float64 = 349.383063 // Mean longitude of the perigee at the epoch
	var mecc float64 = 0.054900      // Eccentricity of the Moon's orbit
	var mangsiz float64 = 0.5181     // Moon's angular size at distance from Earth
	var msmax float64 = 384401       // Semi-major axis of Moon's orbit in km

	moonP.timespace = float64(t.Unix())
	moonP.pdata = utcToJulian(float64(t.Unix()))
	// Calculation of the Sun's position
	var day = moonP.pdata - epoch // Date within epoch

	var n float64 = fixangle((360 / 365.2422) * day) // Mean anomaly of the Sun
	var m float64 = fixangle(n + elonge - elongp)    // Convert from perigee co-orginates to epoch 1980.0
	var ec = kepler(m, eccent)                       // Solve equation of Kepler
	ec = math.Sqrt((1+eccent)/(1-eccent)) * math.Tan(ec/2)
	ec = 2 * rad2deg(math.Atan(ec))               // True anomaly
	var lambdasun float64 = fixangle(ec + elongp) // Sun's geocentric ecliptic longitude

	var f float64 = ((1 + eccent*cos(deg2rad(ec))) / (1 - eccent*eccent)) // Orbital distance factor
	var sunDist float64 = sunsmax / f                                     // Distance to Sun in km
	var sunAng float64 = f * sunangsiz                                    // Sun's angular size in degrees

	// Calsulation of the Moon's position
	var ml float64 = fixangle(13.1763966*day + mmlong)          // Moon's mean longitude
	var mm float64 = fixangle(ml - 0.1114041*day - mmlongp)     // Moon's mean anomaly
	var ev float64 = 1.2739 * sin(deg2rad(2*(ml-lambdasun)-mm)) // Evection
	var ae float64 = 0.1858 * sin(deg2rad(m))                   // Annual equation
	var a3 float64 = 0.37 * sin(deg2rad(m))                     // Correction term
	var mmP float64 = mm + ev - ae - a3                         // Corrected anomaly
	var mec float64 = 6.2886 * sin(deg2rad(mmP))                // Correction for the equation of the centre
	var a4 float64 = 0.214 * sin(deg2rad(2*mmP))                // Another correction term
	var lP float64 = ml + ev + mec - ae + a4                    // Corrected longitude
	var v float64 = 0.6583 * sin(deg2rad(2*(lP-lambdasun)))     // Variation
	var lPP float64 = lP + v                                    // True longitude

	// Calculation of the phase of the Moon
	var moonAge float64 = lPP - lambdasun                   // Age of the Moon in degrees
	var moonPhase float64 = (1 - cos(deg2rad(moonAge))) / 2 // Phase of the Moon

	// Distance of moon from the centre of the Earth
	var moonDist float64 = (msmax * (1 - mecc*mecc)) / (1 + mecc*cos(deg2rad(mmP+mec)))

	var moonDFrac float64 = moonDist / msmax
	var moonAng float64 = mangsiz / moonDFrac // Moon's angular diameter

	// store result
	moonP.phase = fixangle(moonAge) / 360 // Phase (0 to 1)
	moonP.illum = moonPhase               // Illuminated fraction (0 to 1)
	moonP.age = synmonth * moonP.phase    // Age of moon (days)
	moonP.dist = moonDist                 // Distance (kilometres)
	moonP.angdia = moonAng                // Angular diameter (degrees)
	moonP.sundist = sunDist               // Distance to Sun (kilometres)
	moonP.sunangdia = sunAng              // Sun's angular diameter (degrees)
	moonP.phaseHunt()

	moonP.getMoonPosition(t, lat, lng)

	return moonP

}

func (m *Moon) rightAscension(l, b float64) float64 {
	return math.Atan2(sin(l)*cos(e)-math.Tan(b)*sin(e), cos(l))
}

func (m *Moon) declination(l, b float64) float64 {
	return math.Asin(sin(b)*cos(e) + cos(b)*sin(e)*sin(l))
}

func (m *Moon) moonCoords(d float64) (dec, ra, dist float64) { // geocentric ecliptic coordinates of the moon

	L := rad * (218.316 + 13.176396*d) // ecliptic longitude
	M := rad * (134.963 + 13.064993*d) // mean anomaly
	F := rad * (93.272 + 13.229350*d)  // mean distance

	l := L + rad*6.289*sin(M)   // longitude
	b := rad * 5.128 * sin(F)   // latitude
	dt := 385001 - 20905*cos(M) // distance to the moon in km

	ra = m.rightAscension(l, b)
	dec = m.declination(l, b)
	dist = dt
	return
}

func (m *Moon) siderealTime(d float64, lw float64) float64 { return rad*(280.16+360.9856235*d) - lw }

func (m *Moon) calcAzimuth(H, phi, dec float64) float64 {
	return math.Atan2(sin(H), cos(H)*sin(phi)-math.Tan(dec)*cos(phi))
}
func (m *Moon) calcAltitude(H, phi, dec float64) float64 {
	return math.Asin(sin(phi)*sin(dec) + cos(phi)*cos(dec)*cos(H))
}

func astroRefraction(h float64) float64 {
	if h < 0.0 {
		h = 0 // if h = -0.08901179 a div/0 would occur.
	} // the following formula works for positive altitudes only.

	// formula 16.4 of "Astronomical Algorithms" 2nd edition by Jean Meeus (Willmann-Bell, Richmond) 1998.
	// 1.02 / tan(h + 10.26 / (h + 5.10)) h in degrees, result in arc minutes -> converted to rad:
	return 0.0002967 / math.Tan(h+0.00312536/(h+0.08901179))
}

func (m *Moon) getMoonPosition(date time.Time, lat, lng float64) {

	lw := rad * -lng
	phi := rad * lat
	d := toDays(date)

	dec, ra, distance := m.moonCoords(d)
	H := m.siderealTime(d, lw) - ra
	h := m.calcAltitude(H, phi, dec)

	// altitude correction for refraction
	h = astroRefraction(h) // h + rad*0.017/math.Tan(h+rad*10.26/(h+rad*5.10))

	m.azimuth = rad2deg(m.calcAzimuth(H, phi, dec))
	m.altitude = h
	m.distance = distance

	//fmt.Println(m.azimuth, m.altitude, m.distance, m.dist)

}

func integerDivide(a, b int64) (q, r int64) {
	q = a / b
	r = a % b
	if r < 0 {
		q = q - 1
		r = r + b
	}
	return
}

func sin(a float64) float64 {
	return math.Sin(a)
}

func cos(a float64) float64 {
	return math.Cos(a)
}

func rad2deg(r float64) float64 {
	return (r * 180) / math.Pi
}

func deg2rad(d float64) float64 {
	return (d * math.Pi) / 180
}

func fixangle(a float64) float64 {
	return (a - 360*math.Floor(a/360))
}

func kepler(m float64, ecc float64) float64 {
	epsilon := 0.000001
	m = deg2rad(m)
	e := m
	var delta float64
	delta = e - ecc*math.Sin(e) - m
	e -= delta / (1 - ecc*math.Cos(e))
	for math.Abs(delta) > epsilon {
		delta = e - ecc*math.Sin(e) - m
		e -= delta / (1 - ecc*math.Cos(e))
	}
	return e
}

func (m *Moon) phaseHunt() {
	var sdate float64 = utcToJulian(m.timespace)
	var adate float64 = sdate - 45
	var ats float64 = m.timespace - 86400*45
	t := time.Unix(int64(ats), 0)
	var yy float64 = float64(t.Year())
	var mm float64 = float64(t.Month())

	var k1 float64 = math.Floor(float64(yy+((mm-1)*(1/12))-1900) * 12.3685)
	var nt1 float64 = meanPhase(adate, k1)
	adate = nt1
	var nt2, k2 float64

	for {
		adate += synmonth
		k2 = k1 + 1
		nt2 = meanPhase(adate, k2)
		if math.Abs(nt2-sdate) < 0.75 {
			nt2 = truePhase(k2, 0.0)
		}
		if nt1 <= sdate && nt2 > sdate {
			break
		}
		nt1 = nt2
		k1 = k2
	}

	var data [8]float64

	data[0] = truePhase(k1, 0.0)
	data[1] = truePhase(k1, 0.25)
	data[2] = truePhase(k1, 0.5)
	data[3] = truePhase(k1, 0.75)
	data[4] = truePhase(k2, 0.0)
	data[5] = truePhase(k2, 0.25)
	data[6] = truePhase(k2, 0.5)
	data[7] = truePhase(k2, 0.75)

	for i := 0; i < 8; i++ {
		m.quarters[i] = (data[i] - 2440587.5) * 86400 // convert to UNIX time
	}
}

func utcToJulian(t float64) float64 {
	return t/86400 + 2440587.5
}

func julianToUtc(t float64) float64 {
	return t*86400 + 2440587.5
}

func meanPhase(sdate float64, k float64) float64 {
	// Time in Julian centuries from 1900 January 0.5
	var t float64 = (sdate - 2415020.0) / 36525
	var t2 float64 = t * t
	var t3 float64 = t2 * t

	var nt float64
	nt = 2415020.75933 + synmonth*k +
		0.0001178*t2 -
		0.000000155*t3 +
		0.00033*sin(deg2rad(166.56+132.87*t-0.009173*t2))

	return nt
}

func truePhase(k float64, phase float64) float64 {
	k += phase                  // Add phase to new moon time
	var t float64 = k / 1236.85 // Time in Julian centures from 1900 January 0.5
	var t2 float64 = t * t
	var t3 float64 = t2 * t
	var pt float64
	pt = 2415020.75933 + synmonth*k +
		0.0001178*t2 -
		0.000000155*t3 +
		0.00033*sin(deg2rad(166.56+132.87*t-0.009173*t2))

	var m, mprime, f float64
	m = 359.2242 + 29.10535608*k - 0.0000333*t2 - 0.00000347*t3       // Sun's mean anomaly
	mprime = 306.0253 + 385.81691806*k + 0.0107306*t2 + 0.00001236*t3 // Moon's mean anomaly
	f = 21.2964 + 390.67050646*k - 0.0016528*t2 - 0.00000239*t3       // Moon's argument of latitude

	if phase < 0.01 || math.Abs(phase-0.5) < 0.01 {
		// Corrections for New and Full Moon
		pt += (0.1734-0.000393*t)*sin(deg2rad(m)) +
			0.0021*sin(deg2rad(2*m)) -
			0.4068*sin(deg2rad(mprime)) +
			0.0161*sin(deg2rad(2*mprime)) -
			0.0004*sin(deg2rad(3*mprime)) +
			0.0104*sin(deg2rad(2*f)) -
			0.0051*sin(deg2rad(m+mprime)) -
			0.0074*sin(deg2rad(m-mprime)) +
			0.0004*sin(deg2rad(2*f+m)) -
			0.0004*sin(deg2rad(2*f-m)) -
			0.0006*sin(deg2rad(2*f+mprime)) +
			0.0010*sin(deg2rad(2*f-mprime)) +
			0.0005*sin(deg2rad(m+2*mprime))
	} else if math.Abs(phase-0.25) < 0.01 || math.Abs(phase-0.75) < 0.01 {
		pt += (0.1721-0.0004*t)*sin(deg2rad(m)) +
			0.0021*sin(deg2rad(2*m)) -
			0.6280*sin(deg2rad(mprime)) +
			0.0089*sin(deg2rad(2*mprime)) -
			0.0004*sin(deg2rad(3*mprime)) +
			0.0079*sin(deg2rad(2*f)) -
			0.0119*sin(deg2rad(m+mprime)) -
			0.0047*sin(deg2rad(m-mprime)) +
			0.0003*sin(deg2rad(2*f+m)) -
			0.0004*sin(deg2rad(2*f-m)) -
			0.0006*sin(deg2rad(2*f+mprime)) +
			0.0021*sin(deg2rad(2*f-mprime)) +
			0.0003*sin(deg2rad(m+2*mprime)) +
			0.0004*sin(deg2rad(m-2*mprime)) -
			0.0003*sin(deg2rad(2*m+mprime))
		if phase < 0.5 { // First quarter correction
			pt += 0.0028 - 0.0004*cos(deg2rad(m)) + 0.0003*cos(deg2rad(mprime))
		} else { // Last quarter correction
			pt += -0.0028 + 0.0004*cos(deg2rad(m)) - 0.0003*cos(deg2rad(mprime))
		}
	}

	return pt
}

// Phase - returns calculated moon phase
func (m *Moon) Phase() float64 {
	return m.phase
}

// Illumination - returns calculated moon illumination
func (m *Moon) Illumination() float64 {
	return m.illum
}

// Age - returns calculated moon age
func (m *Moon) Age() float64 {
	return m.age
}

// Distance - returns calculated moon distance
func (m *Moon) Distance() float64 {
	return m.dist
}

// Diameter - returns calculated moon diameter
func (m *Moon) Diameter() float64 {
	return m.angdia
}

// SunDistance - returns calculated sun distance
func (m *Moon) SunDistance() float64 {
	return m.sundist
}

// SunDiameter - returns calculated sun diameter
func (m *Moon) SunDiameter() float64 {
	return m.sunangdia
}

// NewMoon - returns calculated new moon (unix) time
func (m *Moon) NewMoon() float64 {
	return m.quarters[0]
}

// FirstQuarter - returns calculated first quarter (unix) time
func (m *Moon) FirstQuarter() float64 {
	return m.quarters[1]
}

// FullMoon - returns calculated full moon (unix) time
func (m *Moon) FullMoon() float64 {
	return m.quarters[2]
}

// LastQuarter - returns calculated last quarter (unix) time
func (m *Moon) LastQuarter() float64 {
	return m.quarters[3]
}

// NextNewMoon - returns calculated next new moon (unix) time
func (m *Moon) NextNewMoon() float64 {
	return m.quarters[4]
}

// NextFirstQuarter - returns calculated next first quarter of moon (unix) time
func (m *Moon) NextFirstQuarter() float64 {
	return m.quarters[1]
}

// NextFullMoon - returns calculated next first full moon (unix) time
func (m *Moon) NextFullMoon() float64 {
	return m.quarters[6]
}

// NextLastQuarter - returns calculated next last quarter of moon (unix) time
func (m *Moon) NextLastQuarter() float64 {
	return m.quarters[7]
}

// PhaseFix - rounding mechanism describing phase range 0-8
func (m *Moon) PhaseFix() int {
	return int(math.Floor((m.phase + 0.0625) * 8))
}

// PhaseName - simple "canned" moon phase names
func (m *Moon) PhaseName() string {
	phase := map[int]string{
		0: "New Moon",
		1: "Waxing Crescent",
		2: "First Quarter",
		3: "Waxing Gibbous",
		4: "Full Moon",
		5: "Waning Gibbous",
		6: "Third Quarter",
		7: "Waning Crescent",
		8: "New Moon",
	}

	return phase[m.PhaseFix()]
}

// PhaseIcon - simple "canned" moon svg, cartoon but recognizable
func (m *Moon) PhaseIcon(sw, sh int) (img draw.Image, err error) {

	// dynamic SVG

	var pos0 float64 = 30
	var sweep [2]int
	var mag float64
	if m.phase <= 0.25 {
		sweep = [2]int{1, 0}
		mag = pos0 - pos0*m.phase*4
	} else if m.phase <= 0.50 {
		sweep = [2]int{0, 0}
		mag = pos0 * (m.phase - 0.25) * 4
	} else if m.phase <= 0.75 {
		sweep = [2]int{1, 1}
		mag = pos0 - pos0*(m.phase-0.50)*4
	} else if m.phase <= 1 {
		sweep = [2]int{0, 1}
		mag = pos0 * (m.phase - 0.75) * 4
	} else {
		return nil, nil
	}

	w, h := 240, 240

	img = image.NewRGBA(image.Rect(0, 0, sw, sh))
	var iconMem = new(bytes.Buffer)

	var canvas = svg.New(iconMem)
	canvas.Start(w, h)

	bfs := svg.Filterspec{In: `SourceGraphic`}
	canvas.Def()
	canvas.Filter(`blur_f`, `filterUnits="objectBoundingBox" x="-50%" y="-50%" width="200%" height="200%"`)
	canvas.FeGaussianBlur(bfs, 6, 6)
	canvas.Fend()
	canvas.Filter(`blur_z`, `filterUnits="objectBoundingBox" x="-50%" y="-50%" width="200%" height="200%"`)
	canvas.FeGaussianBlur(bfs, 2, 2)
	canvas.Fend()
	canvas.DefEnd()

	canvas.Group(fmt.Sprintf("transform=\"rotate(%f %v %v)\"", m.azimuth, w/2, h/2))
	canvas.Group()

	pos1 := `m120,9.5`
	pos2 := `221`

	d := fmt.Sprintf("%[1]sa%[2]v,%.0[4]f 0 1,%[3]d 0,%[5]s ", pos1, mag, sweep[0], pos0, pos2)
	d += fmt.Sprintf("a%.0[2]f,%.0[2]f 0 1,%[1]d 0,-%[3]s", sweep[1], pos0, pos2)
	db := fmt.Sprintf("%[1]sa%.0[2]f,%.0[2]f 0 1 1 0,%[3]sa%.0[2]f,%.0[2]f 0 1 1 0,-%[3]s", pos1, pos0, pos2)
	canvas.Path(db, `style="fill:linen;filter:url(#blur_f);"`)
	canvas.Path(db, `style="fill:#3b5450;fill-opacity:0.8;"`)
	canvas.Path(d, `style="fill:#95bfb9;fill-opacity:0.8;"`)

	canvas.Group(`style="filter:url(#blur_f);"`)

	res := 0.4
	if m.phase > 0.5 && m.phase < 1.1 {
		res = 0.2
	}
	canvas.Circle(70, 70, 26,
		fmt.Sprintf("style=\"fill:#c3e5e0;fill-opacity:%v;\"", res))
	canvas.Path(`m108.27516,210.81657c-1.3504,-10.80513 -6.44946,-11.24919 -18.12884,-16.03732c14.56598,-3.74255 18.02777,-12.64935 12.08516,-23.48683c7.46002,7.73173 14.19221,5.508 22.80197,-8.00359c-5.21947,18.38783 8.57644,22.84642 25.6544,23.38799c-18.60823,6.15093 -23.01136,14.62107 -17.60719,25.04577c-8.50439,-5.66703 -16.28758,-10.25844 -24.80548,-0.90601l-0.00003,0z`,
		fmt.Sprintf("style=\"fill:#c3e5e0;fill-opacity:%v;\"", res))

	canvas.Path(`m128.24406,50.28232c-1.34797,-4.04392 -10.42974,-13.49112 -15.13146,-15.05835c-0.70533,-0.23513 -1.67472,-0.294 -2.92395,-0.21929l0.00012,0c-8.74459,0.52283 -28.86279,7.88339 -32.60208,11.62271c-3.44359,6.13418 -16.05388,15.40858 -18.7864,8.33326c5.00726,-9.49068 8.30148,-14.63581 -10.08763,-10.23383c-13.76045,12.53045 -24.2873,28.53001 -30.18981,46.63706c2.03709,4.44953 1.80735,8.09324 2.41226,10.59934c5.26264,12.40651 0.10338,27.78371 9.35665,36.54941c9.53111,0.17742 8.42476,3.74726 30.55531,3.65494c5.37191,10.7438 -11.12965,14.32599 -7.74848,27.85066c0.43427,1.7371 10.17385,1.40086 12.06131,1.02338c14.02014,-2.80405 6.95739,5.59366 15.47131,7.72214c7.78658,1.94666 12.5189,-12.64263 29.5101,-6.08451c4.28943,-2.85962 -0.75579,-13.5502 -1.70692,-16.40359c-1.83748,-5.51245 3.59325,-17.59399 2.33917,-20.10218c-7.05987,-6.75171 -15.02222,-10.03067 -15.78935,-20.46766c4.75418,-1.86567 9.77763,-2.40728 13.52328,-2.48537c7.79315,-1.55862 -0.73961,-15.35075 8.84497,-15.35075c8.89123,-1.5066 32.49506,5.38301 26.27644,-8.10498c-3.18056,-5.30097 -5.7018,-7.82658 -12.75384,-7.58946c-6.50226,-0.33698 -13.68906,13.52927 -18.12783,8.38456c-9.61082,-6.36477 -20.56831,4.94759 -31.82363,11.03409c-10.32845,-3.18729 -13.36361,-4.14546 -15.03283,-10.74169c5.25304,-5.02206 7.68763,-5.74048 19.3712,-4.67832c5.02433,-0.87899 10.09211,1.75259 15.13146,0.65789c10.54214,-7.2202 21.95662,-9.37039 28.28925,-19.225c2.35979,-5.67731 1.44409,-11.38577 -0.4386,-17.32443l-0.00003,-0.00001zm24.12261,34.57575c1.86193,0 10.32736,-0.47088 11.03792,0.95029c0.89778,1.79554 -17.14001,18.37328 -3.36254,27.55826c2.07507,1.38339 6.13371,2.69804 7.74848,4.31283c-2.38508,20.69419 0.58991,15.32538 6.72509,26.82726c1.4599,2.91979 0.05028,5.52113 1.97366,8.40637c7.86555,11.79832 12.5012,-5.11042 11.11103,-12.06131c-2.32274,-11.61377 -15.80598,4.54151 -15.80598,-13.06805c0,-15.67188 22.82347,-5.11245 22.82347,-20.11881c-7.57119,-0.56495 -1.4425,-12.99641 -3.65494,-15.78935c-2.91191,-13.73771 -12.06507,-13.27288 -22.81065,-8.69495c-0.54537,-7.75473 0.02667,-21.13721 -6.72127,-26.17319c-14.21843,2.79132 -17.41507,-13.24312 -27.04657,4.75143c-3.62162,11.80164 8.85476,22.84174 17.98231,23.09923zm-93.45234,14.4949c-0.64363,3.1706 -4.2577,5.7968 -7.06581,4.65774c-4.20162,-1.70429 -5.21488,-7.2936 -2.12415,-9.63373c3.63306,-2.75076 9.71541,2.38767 9.18996,4.97598l0,0.00001z`, `style="fill:#3b5450;fill-opacity:0.4;"`)
	canvas.Path(`m201.66117,88.85941c2.17407,4.3481 7.51488,12.33959 14.08492,9.05458c5.81756,-8.85652 -3.88549,-17.34894 -9.22227,-19.78592c-2.85099,0.53024 -6.05085,7.92571 -4.86265,10.73135z`, `style="fill:#3b5450;fill-opacity:0.4;"`)
	canvas.Path(`m194.95409,135.13836c1.55746,3.11495 10.19088,8.97975 11.06669,10.73135c2.52515,5.05028 -3.3668,10.71809 -2.01213,12.07277c2.65294,2.65294 8.10239,-5.19198 8.71923,-7.04245c0.7769,-2.33072 5.21906,-16.24362 3.0182,-18.44451c-5.58482,-5.58482 -2.64178,-17.18749 -11.40205,-21.79804c-6.53829,2.87988 -12.82456,17.61167 -9.38994,24.48088z`, `style="fill:#3b5450;fill-opacity:0.4;"`)

	canvas.Gend()
	canvas.Gend()

	canvas.Path(db, `style="fill:cyan;fill-opacity:0.05;filter:url(#blur_z);"`)
	canvas.Path(d, `style="fill:white;fill-opacity:0.1;filter:url(#blur_z);"`)
	canvas.Path(d, `style="fill:paleturquoise;fill-opacity:0.1;filter:url(#blur_z);"`)

	canvas.Gend()

	canvas.End()

	//fmt.Println(m.phase) //iconMem.String())
	iconI, err := oksvg.ReadIconStream(iconMem)
	if err != nil {
		return img, err
	}

	gv := rasterx.NewScannerGV(w, h, img, img.Bounds())
	r := rasterx.NewDasher(w, h, gv)
	iconI.SetTarget(0, 0, float64(sw), float64(sh))
	iconI.Draw(r, 1.0)

	return img, nil

}
