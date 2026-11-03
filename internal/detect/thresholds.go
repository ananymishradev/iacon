package detect

import "time"

const (
	AppReadyOpenBad    = 5
	AppReadyOpenWindow = 6
	AppReadyResolveOK  = 15
	AppReadyResolveAge = 15 * time.Second
	ReadOpenBad        = 3
	ReadResolveOK      = 3
	CanaryOpenBad      = 2
	CanaryResolveOK    = 3
	MergeWindow        = 60 * time.Second
)
