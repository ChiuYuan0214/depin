package depin

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"

	"github.com/ChiuYuan0214/depin/color"
)

// 限制使用時一定要用RegisterScope創建Scope (包括pool的初始化)，所以設定成private type
type scopeType string

func (s scopeType) Match(s2 scopeType) bool {
	return s == s2
}

func (s scopeType) NotEmpty() bool {
	return s.Match(Void) == false
}

const (
	Void   scopeType = ""
	Global scopeType = "Global"
	Test   scopeType = "Test"
	Skip   scopeType = "skip"
)

type Runnable interface {
	Run() error
	Stop()
}

func NewRunnableInfo() (ri *RunnableInfo) {
	ri = new(RunnableInfo)
	ri.Scopes = make(map[scopeType]bool)
	ri.Interfaces = make(map[string]bool)
	return
}

type RunnableInfo struct {
	Name       string
	Scopes     map[scopeType]bool
	Interfaces map[string]bool
}

func (ri *RunnableInfo) GetDescriptor() string {
	var scopeNames []string
	for scope := range ri.Scopes {
		scopeNames = append(scopeNames, string(scope))
	}
	var interNames []string
	for interName := range ri.Interfaces {
		interNames = append(interNames, interName)
	}

	return fmt.Sprintf(
		"implementation %s for interfaces[%s] at scopes[%s].",
		color.Green(ri.Name),
		color.Blue(strings.Join(interNames, ", ")),
		color.Yellow(strings.Join(scopeNames, ", ")),
	)
}

type RunnableSet map[Runnable]*RunnableInfo

func (rs *RunnableSet) Add(name string, runnable Runnable, interfaceName string, scopes []scopeType) {
	info, ok := (*rs)[runnable]
	if !ok {
		info = NewRunnableInfo()
		(*rs)[runnable] = info
	}
	info.Name = name
	info.Interfaces[interfaceName] = true
	for _, scope := range scopes {
		info.Scopes[scope] = true
	}
}

type RunningSet struct {
	mu  sync.Mutex
	set map[Runnable]bool
}

func (rs *RunningSet) Add(runnable Runnable) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.set == nil {
		rs.set = make(map[Runnable]bool)
	}
	rs.set[runnable] = true
}

func (rs *RunningSet) Has(runnable Runnable) bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.set == nil {
		return false
	}
	_, ok := rs.set[runnable]
	return ok
}

func (rs *RunningSet) Remove(runnable Runnable) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.set == nil {
		return
	}
	delete(rs.set, runnable)
}

type Pool map[reflect.Type]any

type Scopes map[scopeType]Pool

func (s *Scopes) Has(scope scopeType) bool {
	_, ok := (*s)[scope]
	return ok
}

func (s *Scopes) Init(scope scopeType) {
	(*s)[scope] = Pool{}
}

func (s *Scopes) Get(scope scopeType) (pool Pool) {
	pool, ok := (*s)[scope]
	if !ok {
		log.Panicf("%s scope not registered", scope)
	}
	return
}

type Tag struct {
	scope scopeType
}
