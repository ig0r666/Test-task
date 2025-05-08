package server

import (
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

type Server struct {
	URL          *url.URL
	status       uint32
	ReverseProxy *httputil.ReverseProxy
}

func NewServer(url *url.URL, proxy *httputil.ReverseProxy) *Server {
	return &Server{
		URL:          url,
		status:       1,
		ReverseProxy: proxy,
	}
}

func (s *Server) GetUrl() *url.URL {
	return s.URL
}

func (s *Server) GetReverseProxy() *httputil.ReverseProxy {
	return s.ReverseProxy
}

// Установить статус серверу, 1 - рабочий, 0 - нет
func (s *Server) SetStatus(status bool) {
	var val uint32 = 0
	if status {
		val = 1
	}
	atomic.StoreUint32(&s.status, val)
}

// Возвращает true если сервер рабочий, иначе false
func (b *Server) IsWorking() bool {
	return atomic.LoadUint32(&b.status) == 1
}
