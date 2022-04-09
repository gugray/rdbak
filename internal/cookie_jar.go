package internal

import (
	"net/http"
	"net/url"
	"sync"
)

type cookieJar struct {
	lock    sync.Mutex
	cookies map[string][]*http.Cookie
}

func newJar() *cookieJar {
	jar := new(cookieJar)
	jar.cookies = make(map[string][]*http.Cookie)
	return jar
}

func (jar *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.lock.Lock()
	defer jar.lock.Unlock()

	// This is not a generically correct solution!
	// In our case, the first time we get the cookies, after login, we're good
	// Later requests to the API apparently return cookies that will make us
	// appear not-authenticated in subsequent calls
	if _, ok := jar.cookies[u.Host]; !ok {
		jar.cookies[u.Host] = cookies
	}
}

func (jar *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies[u.Host]
}
