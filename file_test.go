package relay

import (
	"testing"

	"github.com/influx6/flux"
)

func TestAssetMap(t *testing.T) {
	tree, err := AssetTree("./engine", ".go")

	if err != nil {
		flux.FatalFailed(t, "Unable to create asset map: %s", err.Error())
	}

	if len(tree) <= 0 {
		flux.FatalFailed(t, "expected one key atleast: %s")
	}

	flux.LogPassed(t, "Succesfully created asset map %+s", tree)
}

type dataPack struct {
	Name  string
	Title string
}

func TestTemplateAssets(t *testing.T) {
	tmpl, err := LoadTemplates("./fixtures", ".tmpl", nil, nil)

	if err != nil {
		flux.FatalFailed(t, "Unable to load templates: %+s", err)
	}

	buf := bufPool.Get()
	defer bufPool.Put(buf)

	do := &dataPack{
		Name:  "alex",
		Title: "world war - z",
	}

	err = tmpl.ExecuteTemplate(buf, "base", do)

	if err != nil {
		flux.FatalFailed(t, "Unable to exec templates: %+s", err)
	}

	flux.LogPassed(t, "Loaded Template succesfully: %s", string(buf.Bytes()))
}
