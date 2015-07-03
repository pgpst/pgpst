package version

import (
	"fmt"
	"runtime"
)

const Version = "0.1.0-alpha"

func String(app string) string {
	return fmt.Sprintf("%s v%s (%s)", app, Version, runtime.Version())
}
