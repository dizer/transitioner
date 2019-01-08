package transitioner

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"runtime"
)

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func wrapCallbackError(err error, fn CallbackFunc) error {
	ptr, _, _, _ := runtime.Caller(0)
	file, line := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).FileLine(ptr)
	return errors.Wrap(err, fmt.Sprintf("FSM Callback failed at %s:%d", file, line))
}
