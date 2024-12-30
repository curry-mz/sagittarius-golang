package code

import (
	"code.cd.local/sagittarius/sagittarius-golang/cores/errors"
)

var (
	ParamError              = errors.New(499, "request param error")
	ServerError             = errors.New(500, "server error")
	MessageIllegalError     = errors.New(501, "message illegal error")
	FindTableError          = errors.New(10001, "not find table")
	BetTimeError            = errors.New(10002, "bet time over")
	MaxBetError             = errors.New(10003, "betting exceeds maximum limit")
	AmountInsufficientError = errors.New(10004, "insufficient amount")
	BetFail                 = errors.New(10005, "Betting failed")
)
