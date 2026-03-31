package depin_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/ChiuYuan0214/depin"
)

// interface used to avoid import cycle
// also for the ease of mocking dependencies
type ServiceAInter interface {
	LogA() string
	LogB(logText string) string
}

type ServiceBInter interface {
	LogA() string
	LogB(logText string) string
}

type ServiceA struct {
	serviceB ServiceBInter
}

func (s *ServiceA) Run() (err error) {
	return
}

func (s *ServiceA) Stop() {}

func (s *ServiceA) LogA() string {
	return s.serviceB.LogB("serviceA --> ")
}

func (s *ServiceA) LogB(logText string) string {
	return logText + fmt.Sprintf("serviceA --> address of serviceA = %p, address of serviceB = %p \n", s, s.serviceB)
}

type ServiceB struct {
	serviceA ServiceAInter
}

func (s *ServiceB) Run() (err error) {
	return
}

func (s *ServiceB) Stop() {}

func (s *ServiceB) LogA() string {
	return s.serviceA.LogB("serviceB --> ")
}

func (s *ServiceB) LogB(logText string) string {
	return logText + fmt.Sprintf("serviceB --> address of serviceA = %p, address of serviceB = %p \n", s.serviceA, s)
}

// ability to support bidirectional referencing (better with awareness of the possibility of deadlock)
func Test_Inject(t *testing.T) {
	sA := depin.Set[ServiceAInter](new(ServiceA))
	sB := depin.Set[ServiceBInter](new(ServiceB))

	depin.Run()
	t.Log(sA.LogA())
	t.Log(sB.LogA())
	depin.Stop()
}

type Service interface {
	LogSelf()
	LogDep()
}

type PrivateService struct {
	depService Service `depin:"scope:Global"`
}

func (s *PrivateService) Run() (err error) {
	fmt.Println("private initialized")
	return
}

func (s *PrivateService) Stop() {}

func (s *PrivateService) LogSelf() {
	fmt.Printf("i am %s \n", reflect.Indirect(reflect.ValueOf(s)).Type().Name())
}

func (s *PrivateService) LogDep() {
	fmt.Printf("my dep is %s \n", reflect.Indirect(reflect.ValueOf(s.depService)).Type().Name())
}

type PublicService struct {
	depService Service `depin:"scope:Private"`
}

func (s *PublicService) Run() (err error) {
	fmt.Println("public initialized")
	return
}

func (s *PublicService) Stop() {}

func (s *PublicService) LogSelf() {
	fmt.Printf("i am %s \n", reflect.Indirect(reflect.ValueOf(s)).Type().Name())
}

func (s *PublicService) LogDep() {
	fmt.Printf("my dep is %s \n", reflect.Indirect(reflect.ValueOf(s.depService)).Type().Name())
}

func Test_SpecifyScopeByTag(t *testing.T) {
	Private := depin.CreateScope("Private")
	sA := depin.Set[Service](new(PrivateService), Private)
	sB := depin.Set[Service](new(PublicService))

	depin.Run()

	sA.LogSelf()
	sA.LogDep()
	sB.LogSelf()
	sB.LogDep()

	depin.Stop()
}

func Test_InjectInMultiScope(t *testing.T) {
	Private := depin.CreateScope("Private")
	SubDomainA := depin.CreateScope("SubDomainA")
	SubDomainB := depin.CreateScope("SubDomainB")
	depin.Set[Service](new(PublicService), depin.Global, Private, SubDomainA, SubDomainB)

	depin.Run()
	depin.Stop()
}
