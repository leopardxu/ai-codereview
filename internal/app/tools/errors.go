package tools

import "errors"

var (
    ErrGerritAuth    = errors.New("gerrit auth error")
    ErrRateLimited   = errors.New("rate limited")
    ErrModelTimeout  = errors.New("model timeout")
    ErrInvalidDiff   = errors.New("invalid diff")
)

