package transport

import (
	"sort"
	"strings"

	mid "github.com/yanglunara/discovery/transport/middleware"
)

type Matcher interface {
	Use(mm ...mid.Middleware)
	Add(selector string, mm ...mid.Middleware)
	Match(operation string) []mid.Middleware
}

type matcher struct {
	middleware         map[string][]mid.Middleware
	defaultsMiddleware []mid.Middleware
	prefix             []string
}

func NewMatcher() Matcher {
	return &matcher{
		middleware: make(map[string][]mid.Middleware),
	}
}

func (m *matcher) Use(mm ...mid.Middleware) {
	m.defaultsMiddleware = mm
}

func (m *matcher) Add(selector string, mm ...mid.Middleware) {
	if strings.HasSuffix(selector, "*") {
		selector = strings.TrimSuffix(selector, "*")
		m.prefix = append(m.prefix, selector)
		sort.Slice(m.prefix, func(i, j int) bool {
			return len(m.prefix[i]) > len(m.prefix[j])
		})
	}
	m.middleware[selector] = mm
}

func (m *matcher) Match(operation string) []mid.Middleware {
	ms := make([]mid.Middleware, 0, len(m.defaultsMiddleware))
	if len(m.defaultsMiddleware) > 0 {
		ms = append(ms, m.defaultsMiddleware...)
	}
	if next, ok := m.middleware[operation]; ok {
		ms = append(ms, next...)
	}

	for _, p := range m.prefix {
		if strings.HasPrefix(operation, p) {
			return append(ms, m.middleware[p]...)
		}
	}
	return ms
}
