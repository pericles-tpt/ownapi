package weather

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"
)

type Candidates struct {
	Candidates []Candidate `json:"candidates"`
}

type Candidate struct {
	ID         string `json:"id"`
	Coordinate struct {
		Longitude float64 `json:"longitude"`
		Latitude  float64 `json:"latitude"`
	} `json:"coordinate"`
	Name      string `json:"name"`
	State     string `json:"state"`
	Gridcells struct {
		Forecast struct {
			X int `json:"x"`
			Y int `json:"y"`
		} `json:"forecast"`
		Name  string `json:"name"`
		State string `json:"state"`
	} `json:"gridcells"`
	Postcode struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	} `json:"postcode"`
	Timezone    string `json:"timezone"`
	Type        string `json:"type"`
	FeatureCode string `json:"feature_code"`
	IsAlpine    bool   `json:"is_alpine"`
}

type Place struct {
	Place struct {
		Gridcells struct {
			Forecast struct {
				X int `json:"x"`
				Y int `json:"y"`
			} `json:"forecast"`
		} `json:"gridcells"`
	} `json:"place"`
}

func Format(jsonDaily []uint8, json1Hourly []uint8, json3Hourly []uint8) (string, error) {
	var (
		hourly  ForecastHourly
		hourly3 Forecast3Hourly
		daily   ForecastDaily
	)
	err := json.Unmarshal(jsonDaily, &daily)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(json1Hourly, &hourly)
	if err != nil {
		return "", err
	}
	err = json.Unmarshal(json3Hourly, &hourly3)
	if err != nil {
		return "", err
	}

	var (
		uv     int
		precip int
		min    int
		max    int

		dailyForecast []DisplayDaily
	)
	if len(daily.Forecast.DataPoints) > 0 {
		today := daily.Forecast.DataPoints[0]
		min = int(today.ATM.SurfAir.TempMinCel)
		max = int(today.ATM.SurfAir.TempMaxCel)
		rad := today.ATM.SurfAir.Radiation
		if rad.UV > 0 {
			uv = int(rad.UV)
		}
		precip = int(today.ATM.SurfAir.Precip.Exceeding50PercentChanceTotalMM)

		dailyForecast = make([]DisplayDaily, len(daily.Forecast.DataPoints)-1)
		now := time.Now()
		for i := 1; i < len(daily.Forecast.DataPoints); i++ {
			curr := daily.Forecast.DataPoints[i]
			now = now.Add(24 * time.Hour)
			day := now.Weekday().String()
			display := DisplayDaily{
				Day:     day,
				TempMin: int(curr.ATM.SurfAir.TempMinCel),
				TempMax: int(curr.ATM.SurfAir.TempMaxCel),
			}
			// Not 100% sure this is correct but it's my best guess
			precip := curr.ATM.SurfAir.Precip
			display.PrecipMin = int(math.Ceil(precip.Exceeding75PercentChanceTotalMM))
			display.PrecipMax = int(math.Ceil(precip.Exceeding25PercentChanceTotalMM))

			dailyForecast[i-1] = display
		}
		dailyForecast[0].Day = "tomorrow"
	}

	loc, err := time.LoadLocation(hourly.Meta.LocalTimezone)
	if err != nil {
		return "", err
	}
	// today, err := time.ParseInLocation(time.RFC3339, hourly.Meta.IssueTimeUtc, loc)
	// if err != nil {
	// 	return "", err
	// }
	var (
		n = 5
	)
	parts := make([]string, 0, n+2)
	currTemp := hourly.Forecast[0].DataPoints[0].Atm.SurfAir.TempApparentCel
	parts = append(parts, fmt.Sprintf("TEMP: %.2fc/%dc/%dc (curr/min/max)\n UV: %d\n PRECIP CHANCE: %d%%", currTemp, min, max, uv, precip))
	parts = append(parts, fmt.Sprintf("NEXT %d HOURS", n))
	for i := 0; i < len(hourly.Forecast); i++ {
		f := hourly.Forecast[i]
		j := 0
		for n > 0 {
			if j >= len(f.DataPoints) {
				j = 0
				break
			}

			h := f.DataPoints[j]
			datetime, err := time.ParseInLocation(time.RFC3339, h.TimeUtc, loc)
			if err != nil {
				return "", err
			}
			parts = append(parts, fmt.Sprintf("%s: t: %.1fc, h: %.1f%%, dpt: %.1fc, w: %.1fkm/h", datetime.Local().Format(time.Kitchen), h.Atm.SurfAir.TempCel, h.Atm.SurfAir.HumRelativePercent, h.Atm.SurfAir.TempDewPtCel, mpsToKph(h.Atm.SurfAir.Wind.Speed10mAvgMps)))
			j++
			n--
		}
	}
	return strings.Join(parts, "\n"), nil
}

func ExtractIdAndTimezone(jsonBytes []uint8, postcode string, suburb string, state string) (string, string, error) {
	cands := Candidates{}
	err := json.Unmarshal(jsonBytes, &cands)
	if err != nil {
		return "", "", err
	}

	for _, c := range cands.Candidates {
		if c.Postcode.Name == fmt.Sprint(postcode) && c.Name == suburb && c.State == state {
			return c.ID, c.Timezone, nil
		}
	}
	return "", "", fmt.Errorf("candidate matching postcode: %s, not found", postcode)
}

func ExtractXAndY(jsonBytes []uint8) (int, int, error) {
	place := Place{}
	err := json.Unmarshal(jsonBytes, &place)
	if err != nil {
		return 0, 0, err
	}

	return place.Place.Gridcells.Forecast.X, place.Place.Gridcells.Forecast.Y, nil
}

func mpsToKph(mps float64) float64 {
	mph := mps * 3600.0
	kph := mph / 1000.0
	return kph
}
