package config

import (
	"testing"

	"github.com/pascaldekloe/goe/verify"
)

func TestMerge(t *testing.T) {
	files := []File{
		File{},
		File{
			Bootstrap:    pBool(false),
			Datacenter:   pString("a"),
			Ports:        FilePorts{DNS: pInt(1)},
			JoinAddrsLAN: []string{"a"},
			NodeMeta:     map[string]string{"a": "b"},
		},
		File{
			Bootstrap:    pBool(true),
			Datacenter:   pString("b"),
			Ports:        FilePorts{DNS: pInt(2)},
			JoinAddrsLAN: []string{"b"},
			NodeMeta:     map[string]string{"c": "d"},
		},
		File{},
	}

	got := Merge(files)
	want := File{
		Bootstrap:    pBool(true),
		Datacenter:   pString("b"),
		Ports:        FilePorts{DNS: pInt(2)},
		JoinAddrsLAN: []string{"a", "b"},
		NodeMeta:     map[string]string{"c": "d"},
	}

	if !verify.Values(t, "", got, want) {
		t.FailNow()
	}
}
