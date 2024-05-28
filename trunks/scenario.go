package trunks

type Scenario struct {
	Name     string `yaml:"name"`
	Duration string `yaml:"duration"`
	Accounts uint   `yaml:"accounts"`

	Call       *Action `yaml:"call,omitempty"`
	Transfer   *Action `yaml:"transfer,omitempty"`
	Deposit    *Action `yaml:"deposit,omitempty"`
	Withdrawal *Action `yaml:"withdrawal,omitempty"`
}

type Action struct {
	Pace *Pace `yaml:"pace"`
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
