# CookieJam
A golang library for fetching cookies from browser.

# Noticed

**This library still in progress and should support more browsers in the future. Use it at your own risk.**

Currently, supported browsers:

- Firefox
- Chrome

Supported os:

- Windows

Test on:

- Firefox 77.0.1(64 bit) on windows
- Chrome  85.0(64 bit) on windows


# Quick Start

```go
package main

import (
	"fmt"
	"github.com/IceSandwich/CookieJam"
	"net/http"
	"regexp"
)

func main() {
	client := &http.Client{}

    // Step 1. Create a cookies manager from firefox browser.
    // You can use CookieJam.NewFromChrome if you use chrome browser.
    // Leaving empty parameter means let the program automatically detect the database filename.
    jam, err := CookieJam.NewFromFirefox("")
    if err != nil {
        log.Fatal("Unhandled error when creating instance:", err)
        return
    }

    // Step 2. Filter some cookies we needed.
    // We have all cookies but we just need what we want. Call jam.Filter(...) to do the task.
    // Use jam.Filter() to clear all the filters we set before.
    jam.Filter("github.com", ".github.com")

    // Step 3. Apply filter and fetch cookies from browser.
    if err := jam.FetchCookies(); err != nil {
        log.Fatal("Unable to fetch cookies:", err)
        return
    }

    // Step 4. Create a http request and parse cookies to it.
    req, err := http.NewRequest("GET", "https://github.com/settings/security-log", nil)
    if err != nil {
        log.Fatal("Unable to create a http request:", err)
    }
    req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:77.0) Gecko/20100101 Firefox/77.0")
    jam.LoadToRequest(req) //parse cookies to http request.
    //fmt.Println("Request Cookie:", req.Header.Get("Cookie")) //see what happened after calling `jam.LoadToRequest`.

    // Step 5. Do it!
    resp, err := client.Do(req)
    if err != nil {
        log.Fatal("Unable to get a response:", err)
    }
    if resp.StatusCode != 200 {
        log.Fatal("Server return:", resp.StatusCode)
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Fatal("Unable to read body of response:", body)
    }
    //jam.ParseFromResponse(resp) //save the reponse's cookies. it won't affect the cookies of browser.
    data := string(body)

    // Step 6. Get something you wanted.
    reg := regexp.MustCompile(`Created the repository <a href=".*?">(.*?)</a>`)
    for i, match := range reg.FindAllStringSubmatch(data, -1) {
        fmt.Println("Match[", i, "]:", match[1])
    }
}
```

Output:

```
Match[ 0 ]: IceSandwich/CookieJam
Match[ 1 ]: IceSandwich/materia-theme
Match[ 2 ]: IceSandwich/MarchingSquares
Match[ 3 ]: IceSandwich/AudioToMarkers
Match[ 4 ]: IceSandwich/IceToolbag
Match[ 5 ]: IceSandwich/VSE_Transform_Tools
```

# Install

```bash
go get github.com/IceSandwich/CookieJam
```

