package config

type Config struct {
	Proxy           string
	DisableHeadless bool
	ChromePath      string
	Timeout         int
	Target          string
	TargetList      string
	Targets         map[string]bool
	Output          string
}
