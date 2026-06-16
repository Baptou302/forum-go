package web

import (
	"net/http"
	"net/url"
	"time"
)

type Flash struct {
	Type    string
	Message string
}

const flashCookie = "forum_flash"

func setFlash(w http.ResponseWriter, typ, msg string) {
	http.SetCookie(w, &http.Cookie{
		Name:     flashCookie,
		Value:    url.QueryEscape(typ + "|" + msg),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func consumeFlash(w http.ResponseWriter, r *http.Request) *Flash {
	c, err := r.Cookie(flashCookie)
	if err != nil || c.Value == "" {
		return nil
	}
	http.SetCookie(w, &http.Cookie{
		Name: flashCookie, Value: "", Path: "/", MaxAge: -1, Expires: time.Unix(0, 0),
	})
	raw, err := url.QueryUnescape(c.Value)
	if err != nil {
		return nil
	}
	if i := indexByte(raw, '|'); i >= 0 {
		return &Flash{Type: raw[:i], Message: raw[i+1:]}
	}
	return nil
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}
