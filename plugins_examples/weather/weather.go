package main

import (
	"bufio"
	"fmt"
	forecast "github.com/mlbright/darksky/v2"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

type gps struct {
	name string
	lat  string
	long string
}

var key string

func Scities(cities []gps) string {
	res := []string{}
	for _, city := range cities {
		f, err := forecast.Get(key, city.lat, city.long, "now", forecast.CA, forecast.French)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, fmt.Sprintf("%s %s %dC (%dC) H:%d W:%dkm/h", city.name, f.Currently.Summary, Round(f.Currently.Temperature), Round(f.Currently.ApparentTemperature), Round(f.Currently.Humidity*100), Round(f.Currently.WindSpeed)))
	}
	return strings.Join(res, " | ")
}

func Round(value float64) int {
	if value < 0.0 {
		value -= 0.5
	} else {
		value += 0.5
	}
	return int(value)
}

func main() {
	keybytes, err := ioutil.ReadFile("darksky_key.txt")
	if err != nil {
		log.Fatal(err)
	}
	key = string(keybytes)
	key = strings.TrimSpace(key)

	cities := []gps{
		{"Aigaliers", "44.074622", "4.30553"},
		{"Amsterdam", "52.3745", "4.898"},
		{"Budapest", "47.4984", "19.0405"},
		{"Cournonsec", "43.5482", "3.7"},
		{"Dijon", "47.3216", "5.0415"},
		{"Hanoi", "21.0292", "105.8525"},
		{"Marseille", "43.2962", "5.37"},
		{"Montréal", "45.5088", "-73.554"},
		{"Petrozavodsk", "61.79", "34.39"},
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
				fmt.Printf("PRIVMSG %s :%s\n", par[1], Scities(cities))
			}
		}
	}
}
