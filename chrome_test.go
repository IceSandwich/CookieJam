package CookieJam

import (
	"fmt"
	"net/http"
	"testing"
)

func TestChromeInstance_New_Filter_FetchCookies_LoadToRequest(t *testing.T) {
	/// `New` Test
	ins, err := NewFromChrome("", "")
	if err != nil {
		t.Errorf("NewFromChrome failed. error: %v", err)
		return
	}
	fmt.Printf("dbFn: %s\n", ins.GetDatabaseFileName())
	fmt.Printf("keyLen: %d\n", len(ins.(*chromeInstance).key))
	fmt.Printf("browser: %s\n", ins.GetBrowserName())

	/// `Filter` Test
	ins.Filter(".baidu.com")

	filter := ins.(*chromeInstance).filter
	if filter != "host_key='.baidu.com'" {
		t.Errorf("Filter failed. Got: %s", filter)
		return
	}
	fmt.Println("Filter:", filter)

	/// `FetchCookies` test
	if err := ins.FetchCookies(); err != nil {
		t.Errorf("FetchCookies failed. error: %v", err)
		return
	}

	fmt.Println("Got cookies:", ins.Length())

	for _, cookie := range ins.(*chromeInstance).cookies {
		fmt.Println(cookie.Name + " = " + cookie.Value)
	}

	/// `LoadToRequest` test
	req, err := http.NewRequest("GET", "https://www.baidu.com", nil)
	if err != nil {
		t.Errorf("Unable to create a http request: %v", err)
		return
	}

	ins.LoadToRequest(req)

	fmt.Println("Cookie:", req.Header.Get("Cookie"))
}