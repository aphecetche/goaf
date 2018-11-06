package fstat

import (
	"bytes"
	"encoding/gob"
	"testing"
)

var size int64 = 1234
var path string = "/alice/data/2015/LHC15o/000246087/pass1/15000246087039.9801/AliESDs.FILTER_ESDMUON_WITH_ALIPHYSICS_v5-09-39-01-1.root"
var hostname string = "nansaf02"
var lastmod int64 = 1567891
var lastacc int64 = lastmod + 10

func TestEncode(t *testing.T) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	fi := NewFileInfo(size, path, hostname, lastmod, lastacc)
	err := enc.Encode(fi)
	if err != nil {
		t.Fail()
	}
}

func TestDecode(t *testing.T) {
	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	src := NewFileInfo(size, path, hostname, lastmod, lastacc)
	err := enc.Encode(src)
	dec := gob.NewDecoder(&b)
	var dest FileInfo
	err = dec.Decode(&dest)
	if err != nil {
		t.Errorf("decode failed %s\n", err)
	}
	if dest.Size() != size {
		t.Fail()
	}
	if dest.Host() != hostname {
		t.Fail()
	}
	if dest.Path() != path {
		t.Fail()
	}
	if dest.LastAcc() != lastacc {
		t.Fail()
	}
	if dest.LastMod() != lastmod {
		t.Fail()
	}
}
