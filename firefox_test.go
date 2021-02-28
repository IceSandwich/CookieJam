package CookieJam

import (
	"fmt"
	"net/http"
	"testing"
)

func TestFirefoxInstance_New(t *testing.T) {
	ins, err := NewFromFirefox("")
	if err != nil {
		t.Errorf("NewFromFirefox failed. error: %v", err)
		return
	}
	fmt.Printf("dbFn: %s\n", ins.GetDatabaseFileName())
	fmt.Printf("browser: %s\n", ins.GetBrowserName())
}

func TestFirefoxInstance_Filter(t *testing.T) {
	var ins Jam = &firefoxInstance{
		instance: instance{
			dbFile:  `C:\Users\Cxn\AppData\Roaming/Mozilla/Firefox/Profiles/snmm3anc.default-release/cookies.sqlite`,
			colHost:   "host",
			filter:  "",
			cookies: make([]http.Cookie,0),
		},
	}

	ins.Filter(".baidu.com", "www.baidu.com")

	filter := ins.(*firefoxInstance).filter
	if filter != "host='.baidu.com' or host='www.baidu.com'" {
		t.Errorf("Filter failed. Got: %s", filter)
		return
	}
	fmt.Println("Filter:", filter)
}

func TestFirefoxInstance_FetchCookies_LoadToRequest(t *testing.T) {
	var ins Jam = &firefoxInstance{
		instance: instance{
			dbFile:  `C:\Users\Cxn\AppData\Roaming/Mozilla/Firefox/Profiles/snmm3anc.default-release/cookies.sqlite`,
			colHost: "host",
			filter:  "host='.baidu.com' or host='www.baidu.com'",
			cookies: make([]http.Cookie, 0),
		},
	}

	if err := ins.FetchCookies(); err != nil {
		t.Errorf("FetchCookies failed. error: %v", err)
		return
	}

	fmt.Println("Got cookies:", ins.Length())

	for _, cookie := range ins.(*firefoxInstance).cookies {
		fmt.Println(cookie.Name + " = " + cookie.Value)
	}

	req, err := http.NewRequest("GET", "https://www.baidu.com", nil)
	if err != nil {
		t.Errorf("Unable to create a http request: %v", err)
		return
	}

	ins.LoadToRequest(req)

	fmt.Println("Cookie:", req.Header.Get("Cookie"))
}

