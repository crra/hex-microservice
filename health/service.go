package health

import "time"

type service struct {
	name        string
	version     string
	startupTime time.Time
}

type HealthResult struct {
	Name    string
	Version string
	Uptime  time.Duration
}

type Service interface {
	Health(now time.Time) HealthResult
}

func New(name, version string, startupTime time.Time) Service {
	return &service{
		name:        name,
		version:     version,
		startupTime: startupTime,
	}
}

func (s *service) Health(now time.Time) HealthResult {
	return HealthResult{
		Name:    s.name,
		Version: s.version,
		Uptime:  now.Sub(s.startupTime).Round(time.Second),
	}
}
