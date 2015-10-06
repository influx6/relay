package cli

import (
	"testing"

	"github.com/influx6/flux"
)

var session = NewJSSession("/static/js", nil, true, false)

func TestJSPkgBundler(t *testing.T) {

	js, jsmap, err := session.BuildPkg("github.com/influx6/relay/cli/base", "base")

	if err != nil {
		flux.FatalFailed(t, "Error build gopherjs dir: %s", err)
	}

	if jsmap.Size() > 50 {
		flux.LogPassed(t, "Successfully built js.map package: %d", jsmap.Size())
	}

	if js.Size() > 50 {
		flux.LogPassed(t, "Successfully built js package: %d", js.Size())
	}

}

func TestJSDirBundler(t *testing.T) {
	js, jsmap, err := session.BuildDir("./base", "github.com/influx6/relay/cli/base", "base")

	if err != nil {
		flux.FatalFailed(t, "Error build gopherjs dir: %s", err)
	}

	if jsmap.Size() > 50 {
		flux.LogPassed(t, "Successfully built js.map package: %d", jsmap.Size())
	}

	if js.Size() > 50 {
		flux.LogPassed(t, "Successfully built js package: %d", js.Size())
	}
}
