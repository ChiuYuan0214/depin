package mock_test

import (
	"errors"
	"testing"

	"github.com/ChiuYuan0214/depin/mock"
	"github.com/stretchr/testify/assert"
)

type ServiceInter interface {
	LogA() (string, int, error)
	LogB(logText string) string
}

type MockService struct {
	ctrl *mock.Ctrl[*MockService]
}

func (s *MockService) LogA() (string, int, error) {
	r := s.ctrl.Call("LogA")
	return r[0].(string), r[1].(int), mock.ErrOrNil(r[2])
}

func (s *MockService) LogB(logText string) string {
	return s.ctrl.Call("LogB", logText)[0].(string)
}

func Test_Mock(t *testing.T) {
	mockedService, srvCtrl := mock.Gen[ServiceInter](new(MockService))

	srvCtrl.MockReturnValues("LogA", "Mocked LogA", 0, nil)
	srvCtrl.MockReturnValues("LogB", "Mocked LogB Response")

	logAResult, num, err := mockedService.LogA()
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, logAResult, "Mocked LogA")

	logBResult := mockedService.LogB("test")
	assert.Equal(t, logBResult, "Mocked LogB Response")

	logACallTimes := srvCtrl.CallTimes("LogA")
	logBCallTimes := srvCtrl.CallTimes("LogB")
	assert.Equal(t, logACallTimes, 1)
	assert.Equal(t, logBCallTimes, 1)

	mockErr := errors.New("an error for test")
	srvCtrl.MockReturnValues("LogA", "", 21, mockErr)
	srvCtrl.MockReturnValues("LogB", "nothing to log")

	logAResult, num, err = mockedService.LogA()
	assert.Equal(t, logAResult, "")
	assert.Equal(t, num, 21)
	assert.Equal(t, err, mockErr)

	logBResult = mockedService.LogB("test")
	assert.Equal(t, logBResult, "nothing to log")

	logACallTimes = srvCtrl.CallTimes("LogA")
	logBCallTimes = srvCtrl.CallTimes("LogB")
	assert.Equal(t, logACallTimes, 2)
	assert.Equal(t, logBCallTimes, 2)
}
