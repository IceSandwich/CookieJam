package CookieJam

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type firefoxInstance struct {
	instance
}

func (f *firefoxInstance) FetchCookies() error {
	db, err := sql.Open("sqlite3", f.dbFile)
	if err != nil {
		return errors.New("cannot read database:" + err.Error())
	}
	defer db.Close()

	sqlCmd := "select name, value from moz_cookies"
	if f.filter != "" {
		sqlCmd += " where " + f.filter
	}
	rows, err := db.Query(sqlCmd)
	if err != nil {
		return errors.New(fmt.Sprintf("cannot run sql: >> %s << reason: %s", sqlCmd, err.Error()))
	}
	defer rows.Close()

	var (
		cookieName string
		cookieValue string
	)
	for rows.Next() {
		if err := rows.Scan(&cookieName, &cookieValue); err != nil {
			return errors.New("cannot read data from database:" + err.Error())
		}
		f.cookies = append(f.cookies, http.Cookie{
			Name:       cookieName,
			Value:      cookieValue,
		})
	}

	return nil
}

func (f *firefoxInstance) GetBrowserName() string {
	return "Firefox"
}

func NewFromFirefox(database string) (Jam, error) {
	dbFile := ""
	if database == "" {
		baseDir := path.Join(os.Getenv("APPDATA"), "Mozilla", "Firefox", "Profiles")

		folders, _ := ioutil.ReadDir(baseDir)
		for _, folder := range folders {
			if folder.IsDir() {
				folderTags := strings.Split(folder.Name(), ".")
				if len(folderTags) >= 2 && folderTags[1] == "default-release" {
					dbFile = path.Join(baseDir, folder.Name(), "cookies.sqlite")
					break
				}
			}
		}

		if dbFile == "" {
			return nil, errors.New("cannot get firefox cookie database, please set 'database' argument")
		}
	} else {
		dbFile = database
	}

	if _, err := os.Stat(dbFile); err != nil || os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("cannot access '%s' or this file doesn't exists", dbFile))
	}

	return &firefoxInstance{
		instance: instance{
			dbFile: dbFile,
			colHost:   "host",
			filter:    "",
			cookies:   make([]http.Cookie,0),
		},
	}, nil
}
