package conf

import "time"

type BackendClientConf interface {
	Endpoint() string
	RequestTimeout() time.Duration
}
