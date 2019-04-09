package pool

import "errors"

// Some errors returned by the Put method.
var (
	ErrNilBuffer    = errors.New("cannot put nil-buffer into pool")
	ErrNoPoolBuffer = errors.New("buffer was not taken from pool")
)
