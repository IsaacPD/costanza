package sound

import (
	"io"
)

type Track interface {
	GetReader() (io.Reader, error)
	Start() error
	Stop()
}
