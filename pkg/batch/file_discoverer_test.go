package batch

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func TestFileBatchDiscoverer(t *testing.T) {
	f, err := ioutil.TempFile("", "batch-discover")
	if err != nil {
		t.Errorf("Fail to create temp file")
	}
	defer os.Remove(f.Name())
	if _, err = f.Write([]byte("m1\nm2")); err != nil {
		t.Errorf("Fail to write temp file")
	}
	f.Close()

	d := &FileBatchDiscoverer{Path: f.Name()}
	machines, err := d.Discover()
	if err != nil {
		t.Errorf("Expect no error but got: %+v", err)
	}
	if !reflect.DeepEqual(machines, []string{"m1", "m2"}) {
		t.Errorf("Discovered machines list is not correct: %+v", machines)
	}
}
