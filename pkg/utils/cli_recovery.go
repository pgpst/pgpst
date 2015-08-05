package utils

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"runtime"
)

func reportPanic(w io.Writer, err error) {
	buf := make([]byte, 10000) // 10kB buffer
	bw := runtime.Stack(buf, false)
	stack := bytes.SplitN(buf[:bw], []byte("\n"), 6)

	fmt.Fprintf(
		w,
		"%+v\n\n%s\n",
		err,
		bytes.Join(
			append(stack[0:1], stack[5:]...),
			[]byte("\n"),
		),
	)
}

func CLIRecovery(w io.Writer) {
	if rv := recover(); rv != nil {
		switch cv := rv.(type) {
		case error:
			reportPanic(w, cv)
		case string:
			reportPanic(w, errors.New(fmt.Sprintf("%+v", cv)))
		}
	}
}
