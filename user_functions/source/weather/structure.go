package weather

type ForecastHourly struct {
	Meta struct {
		IssueTimeUtc     string `json:"issue_time_utc"`
		IssueTimeNextUtc string `json:"issue_time_next_utc"`
		LocalTimezone    string `json:"local_timezone"`
	} `json:"meta"`
	Forecast []struct {
		TimeUtc    string `json:"time_utc"`
		DataPoints []struct {
			TimeUtc string `json:"time_utc"`
			Atm     struct {
				SurfAir struct {
					TempApparentCel    float64 `json:"temp_apparent_cel"`
					TempDewPtCel       float64 `json:"temp_dew_pt_cel"`
					TempCel            float64 `json:"temp_cel"`
					HumRelativePercent float64 `json:"hum_relative_percent"`
					Wind               Wind    `json:"wind,omitempty"`
				} `json:"surf_air"`
			} `json:"atm"`
		} `json:"1hourly"`
	} `json:"fcst"`
}

type Wind struct {
	Dirn10mDegT        float64 `json:"dirn_10m_deg_t"`
	Speed10mAvgMps     float64 `json:"speed_10m_avg_mps"`
	Speed10mAvgKts     float64 `json:"speed_10m_avg_kts"`
	GustSpeed10mMaxMps float64 `json:"gust_speed_10m_max_mps"`
	GustSpeed10mMaxKts float64 `json:"gust_speed_10m_max_kts"`
	MixingHeightM      float64 `json:"mixing_height_m"`
}

type Forecast3Hourly struct {
	Meta struct {
		IssueTimeUtc     string `json:"issue_time_utc"`
		IssueTimeNextUtc string `json:"issue_time_next_utc"`
		LocalTimezone    string `json:"local_timezone"`
	} `json:"meta"`
	Forecast []struct {
		TimeUtc    string `json:"time_utc"`
		DataPoints []struct {
			StartTimeUTC string `json:"start_time_utc"`
			EndTimeUTC   string `json:"end_time_utc"`
			ATM          struct {
				SurfAir struct {
					Precip    Precip    `json:"precip,omitempty"`
					Radiation Radiation `json:"radiation,omitempty"`
					Weather   struct {
						IconCode             int     `json:"icon_code"`
						IconRainCode         float64 `json:"icon_rain_code"`
						IconFogCode          float64 `json:"icon_fog_code"`
						IconFrostCode        float64 `json:"icon_frost_code"`
						IconSnowCode         float64 `json:"icon_snow_code"`
						IconThunderstormCode float64 `json:"icon_thunderstorm_code"`
					} `json:"weather"`
					CloudAmtAvgPercent float64 `json:"cloud_amt_avg_percent"`
				} `json:"surf_air"`
			} `json:"atm"`
			Terr struct {
				SurfLand struct {
					FireDanger struct {
						ForestFuelDrynessFactorAvgCode float64 `json:"forest_fuel_dryness_factor_avg_code"`
					} `json:"fire_danger"`
					Snow map[string]interface{} `json:"snow"`
				} `json:"surf_land"`
			} `json:"terr"`
			Ocn struct {
				SurfWater struct {
					Wave struct {
						HeightWindM  *float64 `json:"height_wind_m"`
						TotalHeightM *float64 `json:"total_height_m"`
					} `json:"wave"`
					Swell struct {
						FirstDirnDegT  *float64 `json:"1st_dirn_deg_t"`
						FirstHeightM   *float64 `json:"1st_height_m"`
						SecondDirnDegT *float64 `json:"2nd_dirn_deg_t"`
						SecondHeightM  *float64 `json:"2nd_height_m"`
					} `json:"swell"`
				} `json:"surf_water"`
			} `json:"ocn"`
		} `json:"3hourly"`
	} `json:"fcst"`
}

type ForecastDaily struct {
	Meta struct {
		IssueTimeUtc     string `json:"issue_time_utc"`
		IssueTimeNextUtc string `json:"issue_time_next_utc"`
		LocalTimezone    string `json:"local_timezone"`
	} `json:"meta"`
	Forecast struct {
		DataPoints []struct {
			DateUTC string `json:"date_utc"`
			ATM     struct {
				SurfAir struct {
					TempMaxCel float64 `json:"temp_max_cel"`
					TempMinCel float64 `json:"temp_min_cel"`
					Precip     Precip  `json:"precip,omitempty"`
					Weather    struct {
						IconCode int `json:"icon_code"`
					} `json:"weather"`
					Radiation Radiation `json:"radiation,omitempty"`
				} `json:"surf_air"`
			} `json:"atm"`
			Terr struct {
				SurfLand struct {
					Snow map[string]interface{} `json:"snow"`
				} `json:"surf_land"`
			} `json:"terr"`
			Ocn struct {
				SurfWater struct {
					Sea map[string]interface{} `json:"sea"`
				} `json:"surf_water"`
			} `json:"ocn"`
			Astro struct {
				SunriseUTC *string `json:"sunrise_utc"`
				SunsetUTC  *string `json:"sunset_utc"`
			} `json:"astro"`
		} `json:"daily"`
	} `json:"fcst"`
}

type Precip struct {
	Exceeding10PercentChanceTotalMM float64 `json:"exceeding_10percentchance_total_mm"`
	Exceeding25PercentChanceTotalMM float64 `json:"exceeding_25percentchance_total_mm"`
	Exceeding50PercentChanceTotalMM float64 `json:"exceeding_50percentchance_total_mm"`
	Exceeding75PercentChanceTotalMM float64 `json:"exceeding_75percentchance_total_mm"`
	AnyProbabilityPercent           float64 `json:"any_probability_percent"`
	AnyRestOfDayProbabilityPercent  float64 `json:"any_restofday_probability_percent"`
	TenMMProbabilityPercent         float64 `json:"10mm_probability_percent"`
	TwentyFiveMMProbabilityPercent  float64 `json:"25mm_probability_percent"`
}

type Radiation struct {
	UV          float64 `json:"uv_clear_sky_max_code"`
	PeriodStart string  `json:"uv_period_start"`
	PeriodEnd   string  `json:"uv_period_end"`
}

type ForecastAstro struct {
	Meta struct {
		LocalTimezone string `json:"local_timezone"`
	} `json:"meta"`
	Forecast struct {
		DataPoints []struct {
			DateUTC string `json:"date_utc"`
			Astro   struct {
				SunriseUTC *string `json:"sunrise_utc"`
				SunsetUTC  *string `json:"sunset_utc"`
			} `json:"astro"`
		} `json:"daily"`
	} `json:"fcst"`
}

type DisplayDaily struct {
	Day       string
	TempMin   int
	TempMax   int
	PrecipMin int
	PrecipMax int
}
