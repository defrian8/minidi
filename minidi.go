package minidi

import (
	"fmt"
	"reflect"
	"sync"
)

const (
	// default tag name when used minidi
	defaultTagStruct = "di"
)

type MiniDI interface {
	// SetStructTag set struct tag name if not set default tag name is "di"
	SetStructTag(tag string)
	// Register register service with id and service
	Register(id string, svc interface{}) error
	// MustRegister register service with id and service, if error will panic
	MustRegister(id string, svc interface{})
	// Start all service ordered by register order
	Start() error
	// Stop all service ordered by reverse register order
	Stop()
	// Get service by id
	Get(id string) interface{}
	// Set Custom logger
	SetLogger(logger Logger)
}

type miniDi struct {
	tagStruct string
	mu        sync.RWMutex
	services  map[string]interface{}
	order     []string
	ready     bool
	logger    Logger
}

func New() MiniDI {
	return &miniDi{
		services:  make(map[string]interface{}),
		order:     make([]string, 0),
		logger:    &defaultLogger{},
		tagStruct: defaultTagStruct,
	}
}

func (m *miniDi) SetStructTag(tag string) {
	m.tagStruct = tag
}

func (m *miniDi) SetLogger(logger Logger) {
	m.logger = logger
}

func (m *miniDi) Register(id string, svc interface{}) error {
	if svc == nil {
		return fmt.Errorf("[minidi] cannot register nil service with id '%s'", id)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ready {
		return fmt.Errorf("[minidi] cannot register '%s': container already started", id)
	}

	if _, exists := m.services[id]; exists {
		return fmt.Errorf("[minidi] service '%s' already registered", id)
	}

	m.services[id] = svc
	m.order = append(m.order, id)

	return nil
}

func (m *miniDi) MustRegister(id string, svc interface{}) {
	if err := m.Register(id, svc); err != nil {
		panic(err)
	}
}

func (m *miniDi) Start() error {
	if m.ready {
		return nil
	}

	// inject all needed dependency
	for id, svc := range m.services {
		if err := m.inject(id, svc); err != nil {
			return err
		}
	}

	// start all service ordered by register order
	// if an Object implement OnStart() and OnStop() will called automatically
	for _, id := range m.order {
		svc := m.services[id]
		if s, ok := svc.(Service); ok {
			m.logger.Printf("[minidi] starting up: %s", id)
			if err := s.OnStart(); err != nil {
				return fmt.Errorf("[minidi] start '%s' failed: %w", id, err)
			}
		}
	}

	m.ready = true
	return nil
}

func (m *miniDi) Stop() {
	for i := len(m.order) - 1; i >= 0; i-- {
		id := m.order[i]
		svc := m.services[id]
		if s, ok := svc.(Service); ok {
			m.logger.Printf("[minidi] stopping: %s", id)
			if err := s.OnStop(); err != nil {
				m.logger.Errorf("[minidi] stop '%s' failed: %v", id, err)
			}
		}
	}

	m.ready = false
}

func (m *miniDi) Get(id string) interface{} {
	svc, ok := m.services[id]
	if !ok {
		panic(fmt.Sprintf("[minidi] service '%s' tidak ditemukan", id))
	}
	return svc
}

func (m *miniDi) inject(id string, svc interface{}) error {
	v := reflect.ValueOf(svc)

	// skip if not pointer or struct
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return nil
	}

	elem := v.Elem()
	t := elem.Type()

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		fieldVal := elem.Field(i)

		tag, ok := field.Tag.Lookup(m.tagStruct)
		if !ok {
			continue
		}

		if !fieldVal.CanSet() {
			return fmt.Errorf(
				"[minidi] field '%s.%s' has tag '%s' but not exported",
				id, field.Name, m.tagStruct,
			)
		}

		var resolved interface{}

		// no tag value look by type
		if tag == "" {
			resolved = m.findByType(field.Type)
			if resolved == nil {
				return fmt.Errorf(
					"[minidi] no service with type %s for field '%s.%s'",
					field.Type, id, field.Name,
				)
			}
		} else {
			var exists bool
			resolved, exists = m.services[tag]
			if !exists {
				return fmt.Errorf(
					"[minidi] service '%s' not found (required by '%s.%s')",
					tag, id, field.Name,
				)
			}
		}

		// get dependency value
		resolvedVal := reflect.ValueOf(resolved)
		if !resolvedVal.Type().AssignableTo(field.Type) {
			return fmt.Errorf(
				"[minidi] type not match for '%s.%s': need %s, got %s",
				id, field.Name, field.Type, resolvedVal.Type(),
			)
		}

		// inject dependency to a field struct
		fieldVal.Set(resolvedVal)
	}

	return nil
}

func (m *miniDi) findByType(targetType reflect.Type) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, svc := range m.services {
		if reflect.TypeOf(svc).AssignableTo(targetType) {
			return svc
		}
	}
	return nil
}
