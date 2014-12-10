package state

import (
	"github.com/compose/transporter/pkg/message"
)

type SessionStore interface {
	Save(key, path string, msg *message.Msg) error
	Select(key, path string, msg *message.Msg) error
}
