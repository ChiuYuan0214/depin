package depin

import (
	"fmt"
	"log"
	"reflect"
	"sync"
	"unsafe"
)

// default Scopes
// Test is for unit test mocking
func init() {
	scopes = Scopes{}
	scopes.Init(Global)
	scopes.Init(Test)
}

var tagKey = "depin"
var hasStarted = false
var lock sync.Mutex
var scopes Scopes
var runnableSet = RunnableSet{}
var runningSet = RunningSet{}

// CreateScope 註冊scope到depin中，有註冊的scope才可以透過Set和Equip使用
func CreateScope(name string) scopeType {
	castedScope := scopeType(name)
	scopes.Init(castedScope)
	return castedScope
}

// Reset 清空指定的scope下的所有依賴
func Reset(scope scopeType) {
	if scopes.Has(scope) {
		scopes.Init(scope)
	}
}

// ResetMocks 於test使用，重置所有mock
func ResetMocks() {
	Reset(Test)
}

// equip 會去指定的scope抓取已註冊的impl，並填充到該impl的fields
func equip[T any](impl T, scope scopeType) T {
	pool := scopes.Get(scope)
	implValue := reflect.ValueOf(impl)
	implType := implValue.Type()
	if isPtrToStruct(implValue) == false {
		log.Panicf("impl is not a pointer to struct")
	}
	implValue = implValue.Elem()
	for i := range implValue.NumField() {
		currentPool := pool
		field := implValue.Field(i)
		fieldType := field.Type()
		if isFieldToInject(field) == false {
			continue
		}
		// 如果impl有在field tag指定scope，會強制使用在該scope註冊的impl
		// 否則會在一開始指定的scope下面找impl來注入
		// tag example: `depin:"scope:Global"`
		tag := implValue.Type().Field(i).Tag.Get(tagKey)
		if tag != "" && scope.Match(Test) == false {
			parsedTag := parseTag(tag)
			if parsedTag.scope.Match(Skip) {
				continue
			}
			if parsedTag.scope.NotEmpty() {
				specifiedPool, ok := scopes[parsedTag.scope]
				if !ok {
					log.Panicf("%s scope not registered", parsedTag.scope)
				}
				currentPool = specifiedPool
			}
		}
		dependency, exist := currentPool[fieldType]
		if !exist {
			log.Panicf("dependency %s not found for %s", fieldType, implType)
		}
		depVal := reflect.ValueOf(dependency)
		if field.CanSet() {
			field.Set(depVal)
		} else {
			ptrToField := reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			ptrToField.Set(depVal)
		}
	}
	return impl
}

// Equip 將有註冊到depin的依賴注入到該impl，如果impl有field是interface且在指定的scope中找不到，會panic
// 如果沒有指定scope，會使用Global
func Equip[T any](impl T, scopes ...scopeType) T {
	if hasStarted == false {
		panic("Equip() before Run() is forbidden")
	}
	scope := Global
	if len(scopes) > 0 {
		scope = scopes[0]
	}
	return equip(impl, scope)
}

// EquipMocks 將有註冊到depin的依賴mock注入到該impl
func EquipMocks[T any](impl T) T {
	return equip(impl, Test)
}

// RunAndSet 可預先初始化依賴，使其他依賴在執行 Run() 時就可調用該依賴
func RunAndSet[I any, T Runnable](impl T, scopes ...scopeType) (runnable I) {
	runningSet.Add(impl)
	err := impl.Run()
	if err != nil {
		panic(err)
	}
	return Set[I](impl, scopes...)
}

// Set 會針對指定的interface註冊instance，如果instance並非interface的實作會panic
// 可同時指定多個scope注入同一個instance，如果沒有指定scope，會註冊到Global
func Set[I any, T Runnable](impl T, scopes ...scopeType) (runnable I) {
	if hasStarted {
		panic("set dependency after invoking Run() is forbidden")
	}
	if len(scopes) == 0 {
		scopes = append(scopes, Global)
	}
	for _, scope := range scopes {
		if scope.Match(Test) {
			panic("scope Test is used for mocks")
		}
		runnable = set[I](impl, scope)
	}
	implName := reflect.Indirect(reflect.ValueOf(runnable)).Type().Name()
	interName := reflect.TypeOf((*I)(nil)).Elem().Name()
	runnableSet.Add(implName, impl, interName, scopes)
	return
}

// SetMock 將impl作為mock註冊到depin
func SetMock[I, T any](impl T) I {
	return set[I](impl, Test)
}

func set[I, T any](impl T, scope scopeType) I {
	pool, ok := scopes[scope]
	if !ok {
		log.Panicf("%s scope not registered", scope)
	}
	interType := reflect.TypeOf((*I)(nil)).Elem()
	implType := reflect.TypeOf(impl)
	if implType.Implements(interType) == false {
		log.Panicf("%s does not implements %s", implType, interType)
	}
	pool[interType] = impl
	return any(impl).(I)
}

func Get[I any](scope scopeType) (I, bool) {
	interType := reflect.TypeOf((*I)(nil)).Elem()
	impl, ok := scopes[scope][interType]
	if !ok {
		var none I
		return none, false
	}
	return impl.(I), true
}

func GetGlobal[I any]() (I, bool) {
	return Get[I](Global)
}

// Run 初始化dependencies，ex: init connection
func Run() {
	lock.Lock()
	defer lock.Unlock()
	if hasStarted {
		panic("invoking Run() twice is forbidden")
	}
	for scope, pool := range scopes {
		for _, impl := range pool {
			equip(impl, scope)
		}
	}
	var wg sync.WaitGroup
	for runnable, info := range runnableSet {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if runningSet.Has(runnable) == false {
				if err := runnable.Run(); err != nil {
					log.Panicf("error occurred when initializing implementation [%s]. Error: %s", info.Name, err.Error())
				}
				runningSet.Add(runnable)
			}
			fmt.Println(tagKey + ": initialized " + info.GetDescriptor())
		}()
	}
	wg.Wait()
	hasStarted = true
}

// Stop 關閉dependencies，ex: close connection
func Stop() {
	lock.Lock()
	defer lock.Unlock()
	for runnable := range runnableSet {
		go func() {
			runnable.Stop()
			runningSet.Remove(runnable)
		}()
	}
	hasStarted = false
}
