# minidi

A lightweight dependency injection package for Go using struct tags.

`minidi` simplifies dependency injection by removing repetitive constructor wiring and allowing dependencies to be injected directly into struct fields using tags.

Perfect for small to medium Go applications that want cleaner initialization without adopting a large DI framework.

---

## Features

- Lightweight and minimal
- Struct tag based injection
- Reduces constructor boilerplate
- Easy service registration
- No code generation
- Framework agnostic

## Philosophy

`minidi` is intentionally simple.

It does not try to become a full enterprise DI framework.
The goal is to provide:

- Minimal setup
- Readable code
- Less boilerplate
- Faster development

while still keeping dependencies explicit inside structs.


## Installation

```bash
go get github.com/defrian8/minidi
```

## Motivation

In many Go projects, dependency injection is usually implemented through constructors:

```go
type UserService struct {
    repo UserRepository
    log  Logger
}

func NewUserService(
    repo UserRepository,
    log Logger,
) *UserService {
    return &UserService{
        repo: repo,
        log:  log,
    }
}
```
As projects grow, constructor wiring becomes repetitive and difficult to maintain.

`minidi` provides a simpler approach:

```go
type UserService struct {
    Repo UserRepository `di:"userRepository"`
    Log  Logger         `di:"logger"`
}
```

## Quick Start
### Register Dependencies
```go
container := minidi.New()

container.MustRegister("logger", &Logger{})
container.MustRegister("userRepository", &UserRepository{})

if err := container.Start(); err != nil {
    panic(err)
}

di.Stop()
```
### Example
```go
package main

import (
	"fmt"

	"github.com/defrian8/minidi"
)

type Logger struct{}

func (l *Logger) Info(msg string) {
	fmt.Println("[INFO]", msg)
}

type UserRepository struct{}

func (r *UserRepository) FindAll() []string {
	return []string{"John", "Jane"}
}

type UserService struct {
	Repo *UserRepository `di:"userRepository"`
	Log  *Logger         `di:"logger"`
}

func (s *UserService) GetUsers() []string {
	s.Log.Info("getting users")
	return s.Repo.FindAll()
}

func main() {
	container := minidi.New()

	container.MustRegister("logger", &Logger{})
	container.MustRegister("userRepository", &UserRepository{})
	container.MustRegister("userService", &UserService{})

	if err := container.Start(); err != nil {
		panic(err)
	}

	service := container.Get("userService").(*UserService)

	users := service.GetUsers()

	fmt.Println(users)
}

```
## Real Example
[Simple rest API](https://github.com/defrian8/minidi-example)

## API Reference
### Init minidi

```go
di := minidi.New()
di.SetStructTag("di") // optional, default is "di"
```

### Registration

```go
// Register with explicit ID
err := di.Register("myService", &MyService{})
if err != nil {
    log.Printf("Error occurred: %v", err)
}

// Panic-on-error convenience wrapper
di.MustRegister("myService", &MyService{})
```

### Struct Tag Syntax

```go
type MyService struct {
    DB    *Database   `di:"db"`
    Cache *RedisCache `di:"redis"`
}

func (m *MyService) Foo() {
    fmt.Println("this is foo")
}
```

### Lifecycle Management
```go
type Service interface {
	OnStart() error
	OnStop() error
}
```
`minidi` has auto lifecycle management, each struct that's implement hooks OnStart/OnStop will automatically call when apps is started or stopped. This is **optional**

Example : 

```go
type Sqlite struct {
	*sql.DB
}

func (s *Sqlite) OnStart() error {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	s.DB = db
	return nil
}

func (s *Sqlite) OnStop() error {
	return s.DB.Close()
}
```

### Starting / Stopping
```go
// Start all service
// Automatically call hook OnStart in all services
if err := di.Start(); err != nil {
    panic(err)
}


// Stop all service
// Automatically call hook OnStop
di.Stop()
```

### Get a service
```go
// Get service by ID (panics if not found)
svc := di.Get("myService").(*MyService)
svc.Foo() // call func on MyService
```
## Advanced
## Registering struct impelement interface
```go

type UserRepository interface {
	Create(ctx context.Context, user *User) error
}

type userRepository struct {
	DB *Sqlite `di:"sqlite"`
}

func NewUserRepo() UserRepository {
	return &userRepository{}
}

func (u *userRepository) Create(ctx context.Context, user *User) error {
  _, err := r.DB.ExecContext(ctx, "INSERT ... ", args)
  return err
}

// register repository user
di.MustRegister("userRepository", NewUserRepo())

// usage
type UserService struct {
	Repo UserRepository `di:"userRepository"`
}

```

## Limitations

`minidi` is designed to stay minimal.

It currently focuses on:

- Struct field injection
- Manual registration
- Simple dependency resolution


Injected fields must be exported.

Because `minidi` uses Go reflection, only exported struct fields
(fields starting with uppercase letters) can be injected. This is a limitation of Go reflection, not `minidi`.

✅ Correct:
```go
type UserService struct {
    Repo *UserRepository `di:"userRepository"`
    Log  *Logger         `di:"logger"`
}
```

❌ Invalid:
```go
type UserService struct {
    repo *UserRepository `di:"userRepository"`
    log  *Logger         `di:"logger"`
}
```