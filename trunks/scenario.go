package trunks

import (
	"time"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

type Scenario struct {
	Name string `yaml:"name"`

	Actions []Action `yaml:"actions"`
}

type Action struct {
	Method   string `yaml:"method"`
	Duration string `yaml:"duration"`
	Bridge   string `yaml:"bridge,omitempty"`
	To       string `yaml:"to,omitempty"`
	Pace     *Pace  `yaml:"pace"`
}

func (a *Action) GetPace() vegeta.Pacer {
	if a.Pace.Rate != nil {
		d, _ := time.ParseDuration(a.Pace.Rate.Per)
		return vegeta.Rate{Freq: a.Pace.Rate.Freq, Per: d}
	}
	if a.Pace.Linear != nil {
		d, _ := time.ParseDuration(a.Pace.Linear.Start.Per)
		return vegeta.LinearPacer{
			StartAt: vegeta.ConstantPacer{
				Freq: a.Pace.Linear.Start.Freq,
				Per:  d,
			},
			Slope: a.Pace.Linear.Slope,
		}
	}
	return nil
}

type Pace struct {
	Rate   *PRate   `yaml:"rate,omitempty"`
	Linear *PLinear `yaml:"linear,omitempty"`
}

type PRate struct {
	Freq int    `yaml:"freq"`
	Per  string `yaml:"per"`
}

type PLinear struct {
	Start PRate   `yaml:"start"`
	Slope float64 `yaml:"slope"`
}
