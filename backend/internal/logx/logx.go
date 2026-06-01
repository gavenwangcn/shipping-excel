package logx

import (
	"fmt"
	"log"
	"os"
)

var std = log.New(os.Stdout, "[shipping-excel] ", log.LstdFlags|log.Lmicroseconds)

func Infof(format string, args ...interface{}) {
	std.Printf("INFO  "+format, args...)
}

func Warnf(format string, args ...interface{}) {
	std.Printf("WARN  "+format, args...)
}

func Errorf(format string, args ...interface{}) {
	std.Printf("ERROR "+format, args...)
}

func Jobf(jobID, format string, args ...interface{}) {
	std.Printf("INFO  job=%s "+format, append([]interface{}{jobID}, args...)...)
}

func JobErrf(jobID, format string, args ...interface{}) {
	std.Printf("ERROR job=%s "+format, append([]interface{}{jobID}, args...)...)
}

func HTTP(method, path string, status int, detail string) {
	if detail == "" {
		std.Printf("INFO  http method=%s path=%s status=%d", method, path, status)
		return
	}
	std.Printf("INFO  http method=%s path=%s status=%d detail=%s", method, path, status, detail)
}

func Step(jobID, step string, detail string) {
	if detail == "" {
		Jobf(jobID, "step=%s", step)
		return
	}
	Jobf(jobID, "step=%s %s", step, detail)
}

func Stepf(jobID, step, format string, args ...interface{}) {
	Step(jobID, step, fmt.Sprintf(format, args...))
}
