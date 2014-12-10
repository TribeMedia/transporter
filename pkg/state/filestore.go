package state

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"time"

	"github.com/compose/transporter/pkg/message"
)

type filestore struct {
	filename    string
	flushTicker *time.Ticker
	states      map[string]*msgState
}

type msgState struct {
	Id        string
	Timestamp int64
}

func NewFilestore(filename string, interval time.Duration) *filestore {
	filestore := &filestore{
		filename:    filename,
		flushTicker: time.NewTicker(interval),
	}
	go filestore.startFlusher()
	return filestore
}

func (f *filestore) startFlusher() {
	for _ = range f.flushTicker.C {
		f.flushToDisk()
	}
}

func (f *filestore) flushToDisk() error {
	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	err := enc.Encode(f.states)
	if err != nil {
		return err
	}

	fh, eopen := os.OpenFile(f.filename, os.O_CREATE|os.O_WRONLY, 0666)
	defer fh.Close()
	if eopen != nil {
		return eopen
	}
	n, e := fh.Write(b.Bytes())
	if e != nil {
		return e
	}
	fmt.Fprintf(os.Stderr, "%d bytes successfully written to file\n", n)
	return nil
}

func (f *filestore) Save(key, path string, msg *message.Msg) error {
	f.states[key+"-"+path] = &msgState{Id: msg.IdAsString(), Timestamp: msg.Timestamp}
	return f.flushToDisk()
}

func (f *filestore) Select(key, path string) (string, int64, error) {
	currentState := f.states[key+"-"+path]

	if currentState == nil {
		fh, err := os.Open(f.filename)
		if err != nil {
			return "", 0, err
		}
		states := make(map[string]*msgState)
		dec := gob.NewDecoder(fh)
		err = dec.Decode(&states)
		if err != nil {
			return "", 0, err
		}
		currentState = states[key+"-"+path]
	}
	return currentState.Id, currentState.Timestamp, nil
}
