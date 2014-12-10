package state

// import (
// 	"bytes"
// 	"encoding/binary"

// 	"github.com/siddontang/ledisdb/config"
// 	l "github.com/siddontang/ledisdb/ledis"

// 	"github.com/compose/transporter/pkg/message"
// )

// func NewLedisdb() *ledisdb {
// 	ledis, _ := l.Open(config.NewConfigDefault())
// 	db, _ := ledis.Select(0)
// 	return &ledisdb{
// 		db: db,
// 	}
// }

// type ledisdb struct {
// 	db *l.DB
// }

// func (ledis *ledisdb) Save(key, path string, msg *message.Msg) error {
// 	ts := new(bytes.Buffer)
// 	if err := binary.Write(ts, binary.LittleEndian, msg.Timestamp); err != nil {
// 		return err
// 	}
// 	return ledis.db.HMset(
// 		[]byte(key+"-"+path),
// 		l.FVPair{Field: []byte("id"), Value: []byte(msg.IdAsString())},
// 		l.FVPair{Field: []byte("ts"), Value: ts.Bytes()},
// 	)
// }

// func (ledis *ledisdb) Select(key, path string) (string, int64, error) {
// 	var id string
// 	var ts int64
// 	hash, err := ledis.db.HGetAll([]byte(key + "-" + path))
// 	if err != nil {
// 		return id, ts, err
// 	}
// 	for _, v := range hash {
// 		if string(v.Field) == "id" {
// 			id = string(v.Value)
// 		} else if string(v.Field) == "ts" {
// 			buf := bytes.NewReader(v.Value)
// 			err = binary.Read(buf, binary.LittleEndian, &ts)
// 		}
// 	}
// 	return id, ts, err
// }
