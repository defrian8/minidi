package minidi

// Service is an optional interface for lifecycle management.
type Service interface {
	OnStart() error
	OnStop() error
}
