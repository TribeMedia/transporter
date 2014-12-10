package state

import (
	"reflect"
	"testing"
	"time"

	"github.com/compose/transporter/pkg/message"
)

func TestFilestore(t *testing.T) {
	fs := NewFilestore("/tmp/transporter.db", 10000*time.Millisecond)

	data := []struct {
		key  string
		path string
		id   interface{}
		ts   int64
	}{
		{
			"somelongkey",
			"somepath",
			"123",
			time.Now().Unix(),
		},
		{
			"anoterlongkey",
			"somepath/morepath",
			"1234",
			time.Now().Unix(),
		},
	}

	for _, d := range data {
		err := fs.Save(d.key, d.path, &message.Msg{Id: d.id, Timestamp: d.ts})
		if err != nil {
			t.Errorf("got error: %s\n", err)
			t.FailNow()
		}
	}

	for _, d := range data {
		id, ts, err := fs.Select(d.key, d.path)
		if err != nil {
			t.Errorf("got error: %s\n", err)
			t.FailNow()
		}
		if !reflect.DeepEqual(id, d.id) {
			t.Errorf("wanted: %s, got: %s", d.id, id)
		}
		if !reflect.DeepEqual(ts, d.ts) {
			t.Errorf("wanted: %s, got: %s", d.ts, ts)
		}
	}

}

func TestFilestoreUpdates(t *testing.T) {
	fs := NewFilestore("/tmp/transporter.db", 10000*time.Millisecond)

	data := []struct {
		key  string
		path string
		id   interface{}
		ts   int64
	}{
		{
			"somelongkey",
			"somepath",
			"123",
			time.Now().Unix(),
		},
		{
			"somelongkey",
			"somepath",
			"1234",
			time.Now().Add(10 * time.Second).Unix(),
		},
	}

	for _, d := range data {
		err := fs.Save(d.key, d.path, &message.Msg{Id: d.id, Timestamp: d.ts})
		if err != nil {
			t.Errorf("got error: %s\n", err)
			t.FailNow()
		}
	}

	d := data[len(data)-1]
	id, ts, err := fs.Select(d.key, d.path)
	if err != nil {
		t.Errorf("got error: %s\n", err)
		t.FailNow()
	}
	if !reflect.DeepEqual(id, d.id) {
		t.Errorf("wanted: %s, got: %s", d.id, id)
	}
	if !reflect.DeepEqual(ts, d.ts) {
		t.Errorf("wanted: %s, got: %s", d.ts, ts)
	}

}
