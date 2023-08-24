package connection

import "errors"

var ErrNotConnected = errors.New("not connected")

var ErrIOTimeout = errors.New("i/o timeout")
