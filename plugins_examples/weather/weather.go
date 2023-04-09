package main

import (
	"bufio"
	"fmt"
	"github.com/fridim/cabot/pkg/irc"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"strings"
	// Shortening the import reference name seems to make it a bit easier
	owm "github.com/briandowns/openweathermap"
)

type gps struct {
	name string
	lat  float64
	long float64
}

var key string

func icon(str string, moonPhase float64) string {
	var luna string

	switch {
	case moonPhase >= 0 && moonPhase < 0.12:
		luna = "🌑" // new moon
	case moonPhase >= 0.12 && moonPhase < 0.25:
		luna = "🌒" // waxing cresent
	case moonPhase >= 0.25 && moonPhase < 0.37:
		luna = "🌓" // first quarter
	case moonPhase >= 0.37 && moonPhase < 0.5:
		luna = "🌔" // waxing gibbous
	case moonPhase >= 0.5 && moonPhase < 0.62:
		luna = "🌕" // full moon

	case moonPhase >= 0.62 && moonPhase < 0.75:
		luna = "🌖" // waning gibbous
	case moonPhase >= 0.75 && moonPhase < 0.87:
		luna = "🌗" // last quarter
	case moonPhase >= 0.87 && moonPhase < 1:
		luna = "🌘" // warning crescent
	}

	translation := map[string]string{
		"clear sky":           "☀",
		"clear-night":         luna,
		"rain":                "🌧",
		"snow":                "🌨",
		"sleet":               "❄💧",
		"wind":                "🌬",
		"mist":                "🌫",
		"":                    "☁",
		"few clouds":          "⛅",
		"partly-cloudy-night": "☁" + luna,
		"hail":                "🌨 grêle",
		"thunderstorm":        "⛈",
		"tornado":             "🌪",
	}

	if v, ok := translation[str]; ok {
		return v
	}
	return str
}

func temp(tempf float64) string {
	temp := int(math.Round(tempf))

	/*
		- 00 - White.
		- 01 - Black.
		- 02 - Blue.
		- 03 - Green.
		- 04 - Red.
		- 05 - Brown.
		- 06 - Magenta.
		- 07 - Orange.
		- 08 - Yellow.
		- 09 - Light Green.
		- 10 - Cyan.
		- 11 - Light Cyan.
		- 12 - Light Blue.
		- 13 - Pink.
		- 14 - Grey.
		- 15 - Light Grey.
		- 99 - Default Foreground/Background - Not universally supported.
	*/
	switch {
	case temp < -5:
		// bold blue
		return fmt.Sprintf(
			"%c%c02%d%c", 0x02, 0x03, temp, 0x0f,
		)
	case temp >= -5 && temp <= 5:
		// cyan
		return fmt.Sprintf(
			"%c10%d%c", 0x03, temp, 0x0f,
		)
	case temp > 5 && temp < 30:
		// normal
		return fmt.Sprintf("%d", temp)
	case temp >= 30 && temp < 35:
		// orange
		return fmt.Sprintf(
			"%c07%d%c", 0x03, temp, 0x0f,
		)
	case temp > 35:
		// bold red
		return fmt.Sprintf(
			"%c%c04%d%c", 0x02, 0x03, temp, 0x0f,
		)
	}

	return fmt.Sprintf("%d", temp)
}

func wind(w int) string {
	switch {
	case w > 39:
		// vent frais
		// bold
		return fmt.Sprintf("%c%d%c", 0x02, w, 0x0f)
	case w > 50:
		// Grand frais
		// bold + underline
		return fmt.Sprintf("%c%c%d%c", 0x1f, 0x02, w, 0x0f)
	case w > 62:
		// coupe de vent
		// reverse colors
		return fmt.Sprintf("%c%d%c", 0x16, w, 0x0f)
	case w > 75:
		// reverse colors + bold
		return fmt.Sprintf("%c%c%d%c", 0x16, 0x02, w, 0x0f)
	case w > 89:
		// reverse colors + bold + underline
		return fmt.Sprintf("%c%c%c%d%c", 0x16, 0x02, 0x1f, w, 0x0f)
	}

	return fmt.Sprintf("%d", w)
}

func Scities(cities []gps) string {
	res := []string{}
	w, err := owm.NewCurrent("C", "FR", key)
	if err != nil {
		log.Fatalln(err)
	}
	for _, city := range cities {

		coordinates := owm.Coordinates{
			Longitude: city.long,
			Latitude:  city.lat,
		}

		if err := w.CurrentByCoordinates(&coordinates); err != nil {
			log.Fatalln(err)
		}

		var weather string
		if len(w.Weather) > 0 {
			weather = w.Weather[0].Description
		}

		res = append(
			res,
			fmt.Sprintf(
				"%s %s %sC (%sC) H:%d W:%skm/h",
				city.name,
				weather,
				temp(w.Main.Temp),
				temp(w.Main.FeelsLike),
				w.Main.Humidity,
				wind(int(math.Round(w.Wind.Speed))),
			),
		)
	}
	return strings.Join(res, " | ")
}

func main() {
	keybytes, err := ioutil.ReadFile("owm_api.txt")
	if err != nil {
		log.Fatal(err)
	}
	key = string(keybytes)
	key = strings.TrimSpace(key)

	cities := []gps{
		{"Aigaliers", 44.074622, 4.30553},
		{"Amsterdam", 52.3745, 4.898},
		{"Budapest", 47.4984, 19.0405},
		{"Cournonsec", 43.5482, 3.7},
		{"Dijon", 47.3216, 5.0415},
		{"Hanoi", 21.0292, 105.8525},
		{"Marseille", 43.2962, 5.37},
		{"Montréal", 45.5088, -73.554},
		{"Petrozavodsk", 61.79, 34.39},
	}

	bio := bufio.NewReader(os.Stdin)
	r, _ := regexp.Compile("PRIVMSG (#\\S+) ::meteo")
	for {
		line, err := bio.ReadString('\n')
		if err != nil {
			fmt.Println(err)
		} else {
			par := r.FindStringSubmatch(line)
			if par != nil {
				irc.Privmsg(par[1], Scities(cities))
			}
		}
	}
}
