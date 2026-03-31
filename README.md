# depin

A lightweight dependency injection library for Go, built around interface-based registration with scope support.

## Installation

```bash
go get github.com/ChiuYuan0214/depin
```

## Core Concepts

- **Interface-based injection** — dependencies are registered and resolved by interface type
- **Scopes** — isolate dependency sets (e.g., Global, Private, SubDomainA); built-in `Global` and `Test` scopes
- **Lifecycle** — all registered implementations must satisfy `Runnable` (`Run() error`, `Stop()`); `Run()` initialises them concurrently and `Stop()` tears them down

## Quick Start

```go
// 1. Register implementations before Run()
sA := depin.Set[ServiceAInter](new(ServiceA))
sB := depin.Set[ServiceBInter](new(ServiceB))

// 2. Inject dependencies and start all Runnables
depin.Run()

// 3. Use your services
sA.LogA()

// 4. Tear down
depin.Stop()
```

`Set` automatically wires nil interface fields of the registered struct using other registered implementations in the same scope.

## API Reference

### Registering Dependencies

```go
// Register impl as interface I in the specified scopes (default: Global)
depin.Set[I](impl, scopes...)

// Register and immediately run impl before Run() is called
depin.RunAndSet[I](impl, scopes...)
```

### Lifecycle

```go
depin.Run()   // wires all dependencies and calls Run() on each Runnable concurrently
depin.Stop()  // calls Stop() on all Runnables
```

### Scopes

```go
// Create a named scope (must be created before use)
Private := depin.CreateScope("Private")

// Register into multiple scopes at once
depin.Set[Service](new(MyService), depin.Global, Private, SubDomainA)

// Reset all registrations in a scope
depin.Reset(Private)
```

### Runtime Injection

```go
// Inject dependencies into an existing struct after Run()
instance = depin.Equip[T](instance, scope)
```

### Retrieving Instances

```go
impl, ok := depin.Get[I](scope)
impl, ok := depin.GetGlobal[I]()
```

## Tag-Based Scope Override

A struct field tagged with `depin:"scope:<ScopeName>"` will pull its dependency from that specific scope instead of the default one. Use `depin:"scope:skip"` to exclude a field from injection entirely.

```go
type PublicService struct {
    depService Service `depin:"scope:Private"` // injected from the Private scope
}

type AnotherService struct {
    depService Service `depin:"scope:skip"` // not injected
}
```

## Bidirectional Dependencies

`depin` supports mutual references between services. Be aware of potential deadlocks if `Run()` implementations block on each other.

```go
type ServiceA struct{ serviceB ServiceBInter }
type ServiceB struct{ serviceA ServiceAInter }

sA := depin.Set[ServiceAInter](new(ServiceA))
sB := depin.Set[ServiceBInter](new(ServiceB))
depin.Run() // both are wired correctly
```

## Testing & Mocks

### Registering Mocks

```go
depin.SetMock[ServiceInter](new(MockService))

// Inject mocks into the unit under test
sut = depin.EquipMocks(new(MyService))

// Clean up between tests
depin.ResetMocks()
```

### Mock Controller (`mock.Gen`)

The `mock` sub-package provides a controller to stub return values and assert call counts without any code-generation step.

```go
type MockService struct {
    ctrl *mock.Ctrl[*MockService]
}

func (s *MockService) LogA() (string, int, error) {
    r := s.ctrl.Call("LogA")
    return r[0].(string), r[1].(int), mock.ErrOrNil(r[2])
}

func Test_Example(t *testing.T) {
    mockedService, ctrl := mock.Gen[ServiceInter](new(MockService))

    ctrl.MockReturnValues("LogA", "hello", 42, nil)

    result, num, err := mockedService.LogA()
    // result == "hello", num == 42, err == nil

    assert.Equal(t, 1, ctrl.CallTimes("LogA"))
}
```

#### `mock` Utilities

| Function | Description |
|---|---|
| `mock.Gen[T, I](impl)` | Creates a mock instance and its controller |
| `ctrl.MockReturnValues(method, values...)` | Stubs return values for a method |
| `ctrl.CallTimes(method)` | Returns how many times a method was called |
| `ctrl.Call(method, args...)` | Manually invokes a registered stub |
| `mock.ErrOrNil(val)` | Safely casts a raw value to `error` (nil-safe) |
| `mock.CastOrDefault[T](val)` | Safely casts a raw value to type `T`, returning zero value if nil |

## Built-in Scopes

| Scope | Description |
|---|---|
| `depin.Global` | Default scope for all registrations |
| `depin.Test` | Reserved for mocks; cannot be used with `Set` |
