# Namings

- Conciseness over Brevity: Rather use longer names with better description of the usage of the element instead of making it as short as possible.
- Consistency: Similar elements across the code base should be named by the same name principals.

# Functions

## Method Receivers

Method receivers should be named `t` across all methods.

```go
type Controller struct {}

func (t *Controller) Run() {}
```

## Omitting types in function signatures

Types should never be omitted in function or method signatures.

This should be avoided:

```go
func assembleInfoCard(surName, name, title, address string) string {}
```

Do this instead:

```go
func assembleInfoCard(surName string, name string, title string, address string) string {}
```

## Return values

Return values should be named if it is not clear what a return value is for.

Example:

```go
func getStoreObject(name string) (
  content []byte,
  mimeType string,
  created time.Time,
  err error,
) {}
```

## No naked returns

Go allows the use of naked returns with named return variables. This should be avoided at all.

Do not do this:

```go
func countInSlice(
  v string, s []string,
) (n int) {
  for _, c := range s {
    if c == v {
      n++
    }
  }
  return
}
```

Do this instead:

```go
func countInSlice(
  v string, s []string,
) (n int) {
  for _, c := range s {
    if c == v {
      n++
    }
  }
  return n
}
```

# Structs

## Field order

Fields should optimally be ordered as following:

- Anonymous Fields
- Public Fields
- Private Fields

Also, these categories should be separated by an empty line.

Do not do this:

```go
type Entity struct {
  Id string
}

type Player struct {
  Name   string
  points int
  Age    int
  Entity
}
```

Do this instead:

```go
type Entity struct {
  Id string
}

type Player struct {
  Entity

  Name   string
  Age    int

  points int
}
```

## Initialization

When a struct is initialized, all fields should be named.

Example:

```go
type Player struct {
  Name   string
  Age    int
  points int
}

var user = User{
  Name: "Max",
  Age: 25,
}
```

## Dependency interfaces

Service dependencies are decoupled with interfaces. The consuming service defines the a list of interfaces which only define the methods necessary for the service. These interfaces are defined in an `interfaces.go` in the package of the service.

Example:

```go
// server/interfaces.go

type Database interface {
  ListBooks()
  AddBook()
  DeleteBook()
}

// server/server.go

type Server struct {
  db Database
}

// crawler/interfaces.go

type Database interface {
  AddBook()
}

// crawler/crawler.go

type Crawler struct {
  db Database
}
```

## Callbacks

Callback function types should always be extracted into an extra type which is then used at the site of usage.

Example:

```go
type ErrorCallback func(id int, err error)

type Handler struct {
  onError ErrorCallback
}

func (t *Handler) SetErrorCallback(
  cb ErrorCallback,
) {
  t.onError = cb
}
```

# Channels

Channels should only be used if strictly necessary. When possible, use [semaphore](https://pkg.go.dev/golang.org/x/sync/semaphore) instead of channel structures for worker pool like implementations.

When channels are used in function signatures, their direction should be restricted if possible.

Example:

```go
func sender(c chan<- string) {
	go func() {
		for {
			c <- "hello"
		}
	}()
}

func receiver(c <-chan string) {
	for msg := range c {
		fmt.Println(msg)
	}
}
```
