package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/osu-acm/acm-votes/database"
	"golang.org/x/oauth2"
)

var (
	SessionExpiry = time.Hour * 24 * 3
	email_key     = "auth_email"
	SessionCookie = "session"
)

func GenerateRandomToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func GetEmail(ctx context.Context) string {
	return ctx.Value(email_key).(string)
}

func MustBeAuthenticated(db *sql.DB, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, err := r.Cookie(SessionCookie)
		if err != nil {
			http.Redirect(w, r, "/login?source="+url.QueryEscape(r.URL.Path), http.StatusTemporaryRedirect)
			return
		}
		id := sessionCookie.Value
		email, err := GetSession(db, r.Context(), id)
		if email == "" {
			http.Error(w, "Not logged in properly", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), email_key, email)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func Reject(w http.ResponseWriter, ctx context.Context, code int, reason string, oa2 *oauth2.Config) {
	state := GenerateRandomToken()
	SetAuthCookie(w, state, time.Minute*10)
	w.WriteHeader(code)
	LoginPage(oa2, state, reason).Render(ctx, w)

}

func CallbackHandler(db *sql.DB, oa2 *oauth2.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		destination, err := GetRedirectCookie(r)
		if err != nil {
			destination = "/"
		}

		stateCookie, err := r.Cookie("oauth_state")

		if err != nil || stateCookie.Value != r.URL.Query().Get("state") {
			Reject(w, r.Context(), http.StatusBadRequest, "Mismatched state", oa2)
			return
		}

		tok, err := oa2.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			Reject(w, r.Context(), http.StatusInternalServerError, "Failed to exchange token", oa2)
			return
		}

		email, err := getUserInfo(tok.AccessToken)
		if err != nil {
			Reject(w, r.Context(), http.StatusBadRequest, "Failed to get info", oa2)
			return
		}

		authorized := IsAuthorizedUser(db, r.Context(), email)
		if !authorized {
			Reject(w, r.Context(), http.StatusUnauthorized, fmt.Sprintf("%s unauthorized. Please use university email.", email), oa2)
			return
		}

		session_id, err := CreateSession(db, r.Context(), email)
		http.SetCookie(w, &http.Cookie{
			Name:     SessionCookie,
			Value:    session_id,
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   int(SessionExpiry.Seconds()),
			Path:     "/",
		})

		http.Redirect(w, r, destination, http.StatusTemporaryRedirect)

	})

}

func SetAuthCookie(w http.ResponseWriter, state string, duration time.Duration) {
	ClearAuthCookie(w)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(duration.Seconds()),
	})
}
func ClearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   "oauth_state",
		Value:  "",
		MaxAge: -1,
	})
}

func SetRedirectCookie(w http.ResponseWriter, destination string, duration time.Duration) {
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_source",
		Value:    destination,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(duration.Seconds()),
	})
}
func GetRedirectCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("oauth_source")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
}

func IsAuthorizedUser(db *sql.DB, ctx context.Context, email string) bool {
	queries := database.New(db)
	ok, err := queries.UserExists(ctx, email)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return ok == 1
}
