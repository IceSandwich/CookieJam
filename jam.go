package CookieJam

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Jam interface {
	GetBrowserName() string
	GetDatabaseFileName() string
	Length() int

	Filter(domain ...string)
	FetchCookies() error

	LoadToRequest(r *http.Request)
	ParseFromResponse(response *http.Response)

	LoadToJar(url *url.URL, jar *cookiejar.Jar)
}