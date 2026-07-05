package pagination

import "errors"

var ErrBothCursors = errors.New("after and before cannot be used together")
