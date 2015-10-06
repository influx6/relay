package cli

import (
	"testing"

	"github.com/influx6/flux"
)

func TestBuildConfig(t *testing.T) {
	bo := NewBuildConfig()

	if err := bo.Load("./fixtures/builder.yaml"); err != nil {
		flux.FatalFailed(t, "Unable to load config: %s", err)
	}

	if bo.Addr != ":6000" {
		flux.FatalFailed(t, "Wrong addr value %s expected ':6000'", bo.Addr)
	}

	if bo.Static.Dir != "./static" {
		flux.FatalFailed(t, "Wrong static.dir value %s expected ':6000'", bo.Static.Dir)
	}

	if bo.Package != "bitbucket.org/flow/builder" {
		flux.FatalFailed(t, "Wrong package value %s expected 'bitbucket.org/flow/builder'", bo.Package)
	}

	if bo.ClientPackage != "bitbucket.org/flow/builder/client" {
		flux.FatalFailed(t, "Wrong clientpacket value %s expected 'bitbucket.org/flow/builder/client'", bo.ClientPackage)
	}

	if bo.Name != "builder" {
		flux.FatalFailed(t, "Wrong name value %s expected 'builder'", bo.Addr)
	}
}
