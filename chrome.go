package CookieJam

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/thedevsaddam/gojsonq"
	"net/http"
	"os"
	"path"
	"syscall"
	"unsafe"
)

var (
	dllCrypt32  = syscall.NewLazyDLL("Crypt32.dll")
	dllKernel32 = syscall.NewLazyDLL("Kernel32.dll")

	CryptUnprotectData = dllCrypt32.NewProc("CryptUnprotectData")
	LocalFree = dllKernel32.NewProc("LocalFree")
)

type dataBlob struct {
	cbData uint32
	pbData *byte
}

type chromeInstance struct {
	instance

	key []byte
}

func decryptUnprotectedData(data []byte, length int) ([]byte, error) {
	encryptedBlob := dataBlob{
		cbData: uint32(length),
		pbData: &data[0],
	}
	var decryptedBlob dataBlob
	ret, _, err := CryptUnprotectData.Call(uintptr(unsafe.Pointer(&encryptedBlob)), 0, 0, 0, 0, 1, uintptr(unsafe.Pointer(&decryptedBlob)))
	if ret == 0 {
		return nil, err
	}

	decrypted := make([]byte, decryptedBlob.cbData)
	copy(decrypted, (*[1 << 30]byte)(unsafe.Pointer(decryptedBlob.pbData))[:])

	LocalFree.Call(uintptr(unsafe.Pointer(decryptedBlob.pbData)))

	return decrypted, nil
}

func decryptChromeKey(key string) ([]byte, error) {
	encryptedKey := make([]byte, 512)
	lenKey, err := base64.StdEncoding.Decode(encryptedKey, []byte(key))
	if err != nil {
		return nil, err
	}
	//println("PreTag:", string(encryptedKey[:5])) //PreTag: DPAPI

	data, err := decryptUnprotectedData(encryptedKey[5:], lenKey-5)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func decryptChromeValue(key []byte, data []byte) (string, error) {
	oldPreTag := []byte{0x01, 0x00, 0x00, 0x00, 0xD0, 0x8C, 0x9D, 0xDF, 0x01, 0x15, 0xD1, 0x11, 0x8C, 0x7A, 0x00, 0xC0, 0x4F, 0xC2, 0x97, 0xEB}
	newPreTag := []byte{'v', '1', '0'}

	if bytes.Equal(data[:len(newPreTag)], newPreTag) {
		block, err := aes.NewCipher(key)
		if err != nil {
			return "", err
		}

		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return "", err
		}

		nonceSize := gcm.NonceSize()
		nonce := data[3 : 3+nonceSize]
		ciphertext := data[3+nonceSize:]

		plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return "", err
		}

		return string(plaintext), nil
	} else if bytes.Equal(data[:len(oldPreTag)], oldPreTag) {
		ret, err := decryptUnprotectedData(data, len(data))
		if err != nil {
			return "", err
		}

		return string(ret), nil
	}

	return "", errors.New("unknown version of cipher")
}

func (f *chromeInstance) FetchCookies() error {
	db, err := sql.Open("sqlite3", f.dbFile)
	if err != nil {
		return errors.New("cannot read database:" + err.Error())
	}
	defer db.Close()

	sqlCmd := "select name, encrypted_value from cookies"
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
		cookieEncrypted []byte
		cookieValue string
	)
	for rows.Next() {
		if err := rows.Scan(&cookieName, &cookieEncrypted); err != nil {
			return errors.New("cannot read data from database:" + err.Error())
		}
		cookieValue, err = decryptChromeValue(f.key, cookieEncrypted)
		if err != nil {
			return errors.New("cannot decrypt data for cookie " + cookieName + " :" + err.Error())
		}
		f.cookies = append(f.cookies, http.Cookie{
			Name:       cookieName,
			Value:      cookieValue,
		})
	}

	return nil
}

func (f *chromeInstance) GetBrowserName() string {
	return "Chrome"
}

func NewFromChrome(database string, localState string) (Jam, error) {
	dbFile := ""
	if database == "" {
		dbFile = path.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data", "Default", "Cookies")
	} else {
		dbFile = database
	}

	if _, err := os.Stat(dbFile); err != nil || os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("cannot access '%s' or this file doesn't exists", dbFile))
	}

	stateFile := ""
	if localState == "" {
		stateFile = path.Join(os.Getenv("LOCALAPPDATA"), "Google", "Chrome", "User Data", "Local State")
	} else {
		stateFile = localState
	}

	if _, err := os.Stat(stateFile); err != nil || os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("cannot access '%s' or this file doesn't exists", stateFile))
	}

	encryptedKey := gojsonq.New().File(stateFile).From("os_crypt.encrypted_key").Get().(string)

	if encryptedKey == "" {
		return nil, errors.New("cannot get encrypted_key from local state file")
	}

	key, err := decryptChromeKey(encryptedKey)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot decrypt chrome key: %s", err.Error()))
	}

	return &chromeInstance{
		instance: instance{
			colHost: "host_key",
			dbFile:  dbFile,
			filter:  "",
			cookies: make([]http.Cookie, 0),
		},
		key: key,
	}, nil
}