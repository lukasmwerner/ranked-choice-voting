package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/osu-acm/acm-votes/auth"
	"github.com/osu-acm/acm-votes/pages"
	"github.com/templui/templui/assets"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

func Main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	oa2 := &oauth2.Config{}
	oa2.Endpoint = endpoints.Google
	oa2.Scopes = []string{"https://www.googleapis.com/auth/userinfo.email"}
	oa2.ClientID = os.Getenv("GOOGLE_KEY")
	oa2.ClientSecret = os.Getenv("GOOGLE_SECRET")
	oa2.RedirectURL = "http://localhost:8080/api/auth/callback/google"

	db, err := sql.Open("sqlite3", "data/data.db")
	if err != nil {
		log.Println("error opening database", err.Error())
		return
	}

	http.Handle("/api/auth/callback/google", auth.CallbackHandler(db, oa2))
	http.Handle("POST /vote/{ballot}", auth.MustBeAuthenticated(db, func(w http.ResponseWriter, r *http.Request) {

	}))
	http.Handle("/vote/{ballot}", auth.MustBeAuthenticated(db, func(w http.ResponseWriter, r *http.Request) {
		people := make([]string, 4)
		for i := range 4 {
			people[i] = faker.Name()
		}

		pages.VotingPage("ACM President", people).Render(r.Context(), w)
	}))
	setupAssetsRoutes(http.DefaultServeMux)
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		state := auth.GenerateRandomToken()

		if cookie, err := r.Cookie("oauth_state"); err == nil && cookie.Value != "" {
			fmt.Println("keeping cookie")
			state = cookie.Value
		} else {
			auth.SetAuthCookie(w, state, time.Minute*10)
		}

		if r.URL.Query().Has("source") {
			auth.SetRedirectCookie(w, r.URL.Query().Get("source"), time.Minute*10)
		}

		auth.LoginPage(oa2, state, "").Render(r.Context(), w)
	})
	http.HandleFunc("/logout", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:     auth.SessionCookie,
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
			MaxAge:   -1,
			Path:     "/",
		})

		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
	})
	log.Println("Listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func setupAssetsRoutes(mux *http.ServeMux) {
	isDevelopment := os.Getenv("GO_ENV") != "production"

	assetHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isDevelopment {
			w.Header().Set("Cache-Control", "no-store")
		} else {
			w.Header().Set("Cache-Control", "public, max-age=31536000")
		}

		var fs http.Handler
		if isDevelopment {
			fs = http.FileServer(http.Dir("./assets"))
		} else {
			fs = http.FileServer(http.FS(assets.Assets))
		}

		fs.ServeHTTP(w, r)
	})

	mux.Handle("GET /assets/", http.StripPrefix("/assets/", assetHandler))
}
