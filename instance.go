package CookieJam

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type instance struct {
	colHost string

	dbFile string
	filter string
	cookies []http.Cookie
}

func (i *instance) Filter(domain ...string) {
	i.filter = ""
	if len(domain) == 0 { // clear filter
		return
	}

	for _, name := range domain {
		i.filter += i.colHost + "='" + name + "' or "
	}
	i.filter = i.filter[:len(i.filter)-len(" or ")]
}

func (i *instance) LoadToJar(url *url.URL, jar *cookiejar.Jar) {
	cookiePointers := make([]*http.Cookie, len(i.cookies))
	for i, cookie := range i.cookies {
		cookiePointers[i] = &cookie
	}
	jar.SetCookies(url, cookiePointers)
}

func (i *instance) Length() int {
	return len(i.cookies)
}

func (i *instance) LoadToRequest(r *http.Request) {
	for _, item := range i.cookies {
		r.AddCookie(&item)
	}
}

func (i *instance) ParseFromResponse(response *http.Response) {
	for _, cookie := range response.Cookies() {
		i.cookies = append(i.cookies, *cookie)
	}
}

func (i *instance) GetDatabaseFileName() string {
	return i.dbFile
}