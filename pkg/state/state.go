package state

import (
	"github.com/compose/transporter/pkg/message"
)

type SessionStore interface {
	Save(path string, msg *message.Msg) error
	Select(path string, msg *message.Msg) error
}
