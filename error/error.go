package error

import (
	"fmt"
	"log"
	"runtime"
	"strings"
)

type Op string
type ECode struct {
	Code int
	Kind string
}

func getMethodName(depthList ...int) string {
	var depth int
	if depthList == nil {
		depth = 1
	} else {
		depth = depthList[0]
	}
	function, _, _, _ := runtime.Caller(depth)
	name := runtime.FuncForPC(function).Name()
	names := strings.Split(name, "/")
	return names[len(names)-1]
}

// Error defines a standard application error.
type Error struct {
	Code     int
	Kind     string
	Op       Op
	Location string
	Err      error
	Message  string
}

func (s *Error) Report() {
	fmt.Println("abc")
}

type AdvError struct {
	Error
	Data string
}

func (s *AdvError) Report() {
	fmt.Println("bcd")
}

func E(args ...interface{}) Error {
	_, file, line, _ := runtime.Caller(1)
	fmt.Println()
	if len(args) == 0 {
		panic("call to errors.E with no arguments")
	}
	e := &Error{}
	for _, arg := range args {
		switch arg := arg.(type) {
		case Op:
			e.Op = arg
		case ECode:
			e.Code = arg.Code
			e.Kind = arg.Kind
		case *Error:
			// Make a copy
			copy := *arg
			e.Err = &copy
		case error:
			e.Err = arg
		case string:
			e.Message = arg
		default:

			log.Printf("errors.E: bad call from %s:%d: %v", file, line, args)
		}
	}
	e.Location = fmt.Sprintf("%s:%d", file, line)
	if e.Op == "" {
		e.Op = Op(getMethodName(2))
	}
	return e
}

func (e *Error) Error() (res string) {
	res = string(e.Op)
	if e.Location != "" {
		names := strings.Split(e.Location, "/")
		res += "(" + names[len(names)-1] + ")"
	}
	if e.Kind != "" {
		res = res + " " + string(e.Kind) + ": \""
	}
	res = res + (e.Message + "\"")
	if e.Err != nil {
		res += e.Err.Error()
	}
	return "\n\tERR:->\t" + res
}
