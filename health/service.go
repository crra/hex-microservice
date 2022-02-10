package health

type service struct {
	name    string
	version string
}

type HealthResult struct {
	Name       string
	Version    string
	Repository bool
}

type Service interface {
	Health() HealthResult
}

func New(name, version string) Service {
	return &service{
		name:    name,
		version: version,
	}
}

func (s *service) Health() HealthResult {
	return HealthResult{
		Name:    s.name,
		Version: s.version,
	}
}
