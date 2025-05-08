package core

import (
	"net/http/httputil"
	"net/url"
)

type Server interface {
	SetStatus(bool)
	IsWorking() bool
	GetUrl() *url.URL
	GetReverseProxy() *httputil.ReverseProxy
}

// Благодаря использованию интерфейсов предусмотрена возможность замены алгоритма балансировщика
// Для использования балансировщика с алгоритмом least connections (или другим) достаточно будет реализовать каждый из методов интерфейса
type Pooler interface {
	AddServer(Server)
	GetNextIndex() int
	ChangeServerStatus(*url.URL, bool)
	GetNextServer() Server
	HealthCheck()
	IsServerWorking(*url.URL) bool
}
