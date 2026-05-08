package minidi_test

import (
	"fmt"
	"testing"

	"github.com/defrian8/minidi"
)

type Logger struct{ prefix string }

func NewLogger() *Logger         { return &Logger{prefix: "[APP]"} }
func (l *Logger) Log(msg string) { fmt.Printf("%s %s\n", l.prefix, msg) }
func (l *Logger) OnStart() error { l.Log("Logger started"); return nil }
func (l *Logger) OnStop() error  { l.Log("Logger stopped"); return nil }

type Database struct {
	Logger *Logger `di:"logger"`
	dsn    string
}

func NewDatabase() *Database { return &Database{dsn: "postgres://localhost/app"} }

func (d *Database) OnStart() error {
	d.Logger.Log(fmt.Sprintf("Database connected: %s", d.dsn))
	return nil
}

func (d *Database) OnStop() error         { d.Logger.Log("Database disconnected"); return nil }
func (d *Database) Query(q string) string { return fmt.Sprintf("result of: %s", q) }

type Cache struct {
	Logger *Logger `di:""`
}

func NewCache() *Cache                 { return &Cache{} }
func (c *Cache) OnStart() error        { c.Logger.Log("Cache started"); return nil }
func (c *Cache) OnStop() error         { c.Logger.Log("Cache stopped"); return nil }
func (c *Cache) Get(key string) string { return fmt.Sprintf("cached:%s", key) }

type UserService struct {
	DB    *Database `di:"database"`
	Cache *Cache    `di:"cache"`
	Log   *Logger   `di:"logger"`
}

func NewUserService() *UserService { return &UserService{} }
func (u *UserService) GetUser(id string) string {
	cached := u.Cache.Get(id)
	u.Log.Log(fmt.Sprintf("GetUser(%s) → %s", id, cached))
	return cached
}

type MyLogger interface {
	Printf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
}

type myLogger struct{}

func (l *myLogger) Printf(format string, args ...interface{}) {}
func (l *myLogger) Errorf(format string, args ...interface{}) {}

func TestSimpleDI(t *testing.T) {
	m := minidi.New()
	m.MustRegister("logger", NewLogger())
	m.MustRegister("database", NewDatabase())

	if err := m.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer m.Stop()

	db := m.Get("database").(*Database)
	if db.Logger == nil {
		t.Fatal("Logger is nil")
	}

	result := db.Query("SELECT 1")
	if result == "" {
		t.Fatal("Query failed")
	}

	fmt.Println("TestSimpleDI OK:", result)
}

func TestTypeBasedInject(t *testing.T) {
	m := minidi.New()
	m.MustRegister("logger", NewLogger())
	m.MustRegister("cache", NewCache())

	if err := m.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer m.Stop()

	cache := m.Get("cache").(*Cache)
	if cache.Logger == nil {
		t.Fatal("Logger is nil")
	}
	fmt.Println("TestTypeBasedInject OK:", cache.Get("user:1"))
}

func TestMultiDependency(t *testing.T) {
	m := minidi.New()
	m.MustRegister("logger", NewLogger())
	m.MustRegister("database", NewDatabase())
	m.MustRegister("cache", NewCache())
	m.MustRegister("userService", NewUserService())

	if err := m.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer m.Stop()

	us := m.Get("userService").(*UserService)
	if us.DB == nil || us.Cache == nil || us.Log == nil {
		t.Fatal("Dependency UserService is nil")
	}
	result := us.GetUser("42")
	fmt.Println("TestMultiDependency OK:", result)
}

func TestGet(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Error("expected panic")
		}
	}()

	m := minidi.New()
	m.MustRegister("logger", NewLogger())
	if err := m.Start(); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer m.Stop()

	if m.Get("logger") == nil {
		t.Fatal("Get should return existing service")
	}
	if m.Get("unknownservice") != nil {
		t.Fatal("Get should return nil for non-existent service")
	}
	fmt.Println("TestGet OK")
}

func TestMissingDependency(t *testing.T) {
	m := minidi.New()
	// Database butuh "logger" tapi tidak didaftarkan
	m.MustRegister("database", NewDatabase())

	err := m.Start()
	if err == nil {
		t.Fatal("Should error because logger is not registered")
	}
	fmt.Println("TestMissingDependency OK (error expected):", err)
}

func TestStartIdempotent(t *testing.T) {
	m := minidi.New()
	m.MustRegister("logger", NewLogger())

	if err := m.Start(); err != nil {
		t.Fatal(err)
	}
	// Call Start() again — should not panic or double-startup
	if err := m.Start(); err != nil {
		t.Fatal("Second Start() should be no-op")
	}
	fmt.Println("TestStartIdempotent OK")
}

func TestSetStruct(t *testing.T) {
	m := minidi.New()
	m.SetStructTag("di")

}

func TestSetLogger(t *testing.T) {
	m := minidi.New()

	defaultLogger := &myLogger{}
	m.SetLogger(defaultLogger)
}
