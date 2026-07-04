package bridge

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const IsFreeBasicsKey contextKey = "is_freebasics"

func DetectFreeBasics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isFB := false

		if r.Header.Get("X-IORG-FBS") == "true" {
			isFB = true
		}
		if !isFB {
			via := r.Header.Get("Via")
			if strings.Contains(via, "Internet.org") {
				isFB = true
			}
		}
		if !isFB {
			ua := r.Header.Get("User-Agent")
			if strings.Contains(ua, "InternetOrgApp") {
				isFB = true
			}
		}
		if !isFB {
			if r.URL.Query().Get("__fb") == "1" {
				isFB = true
			}
		}

		if isFB {
			w.Header().Set("X-Robots-Tag", "noindex")
		}

		ctx := context.WithValue(r.Context(), IsFreeBasicsKey, isFB)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// SimulateFreeBasics sets the X-IORG-FBS header on a request for testing.
func SimulateFreeBasics(r *http.Request) *http.Request {
	r.Header.Set("X-IORG-FBS", "true")
	return r
}

func IsFreeBasics(r *http.Request) bool {
	v, _ := r.Context().Value(IsFreeBasicsKey).(bool)
	return v
}
