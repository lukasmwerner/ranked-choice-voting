package server

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/osu-acm/acm-votes/auth"
	"github.com/osu-acm/acm-votes/database"
	"github.com/osu-acm/acm-votes/pages"
	rankedchoice "github.com/osu-acm/acm-votes/ranked_choice"
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
	oa2.RedirectURL = "https://polling.pdx.land/api/auth/callback/google"
	//oa2.RedirectURL = "http://localhost:8080/api/auth/callback/google"

	db, err := sql.Open("sqlite3", "data/data.db")
	if err != nil {
		log.Println("error opening database", err.Error())
		return
	}
	queries := database.New(db)
	admins := strings.Split(os.Getenv("ADMINS"), ",")

	http.Handle("/admin", auth.MustBeSpecificUser(db, admins, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ballots, err := queries.GetAllBallots(ctx)
		if err != nil {
			http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		candidates, err := queries.GetAllCandidates(ctx)
		if err != nil {
			http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		pages.AdminPage("", ballots, candidates).Render(ctx, w)
	}))

	http.Handle("POST /admin/ballot/tally", auth.MustBeSpecificUser(db, admins, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		r.ParseForm()
		ballot := r.Form.Get("ballot")
		ballotID, err := uuid.Parse(ballot)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		winner, err := rankedchoice.RunBallot(db, ballotID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		name, err := queries.GetCandidateName(ctx, winner)
		if err != nil {
			http.Error(w, fmt.Sprintf("err: %s id: %s", err.Error(), winner.String()), http.StatusInternalServerError)
			return
		}
		ballotName, _ := queries.GetBallotName(ctx, ballotID)

		fmt.Fprintf(w, "Winner of %s: %s", ballotName, name)

	}))
	http.Handle("POST /admin/ballot/change", auth.MustBeSpecificUser(db, admins, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		r.ParseForm()
		status := r.Form.Get("status")
		ballot := r.Form.Get("ballot")
		ballotID, err := uuid.Parse(ballot)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		switch status {
		case "open":
			queries.OpenBallot(ctx, ballotID)
		case "close":
			queries.CloseBallot(ctx, ballotID)
		default:
			http.Error(w, "invalid request", http.StatusBadRequest)
		}

		http.Redirect(w, r, "/admin", http.StatusSeeOther)

	}))

	http.Handle("/api/auth/callback/google", auth.CallbackHandler(db, oa2))
	http.Handle("POST /vote/{ballot}", auth.MustBeAuthenticated(db, func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ballotID, err := uuid.Parse(r.PathValue("ballot"))
		if err != nil {
			http.Error(w, "invalid ballot id", http.StatusBadRequest)
			return
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "unable to read form", http.StatusBadRequest)
			return
		}
		choices := r.PostForm["rank[]"]
		voterID := auth.GetEmail(ctx)

		status, err := queries.GetBallotStatus(ctx, ballotID)
		if err != nil {
			http.Error(w, "ballot not found", http.StatusNotFound)
			return
		}
		if status != "open" {
			http.Error(w, "ballot is not open", http.StatusForbidden)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		qtx := queries.WithTx(tx)
		if err := qtx.DeleteVotes(ctx, database.DeleteVotesParams{
			Email:    voterID,
			BallotID: ballotID,
		}); err != nil {
			tx.Rollback()
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		for rank, candidate := range choices {
			candidateID, _ := uuid.Parse(candidate)
			if err := qtx.InsertRankedVote(ctx, database.InsertRankedVoteParams{
				Email:       voterID,
				BallotID:    ballotID,
				CandidateID: candidateID,
				Rank:        int64(rank + 1),
			}); err != nil {
				tx.Rollback()
				http.Error(w, "internal error", http.StatusInternalServerError)
				return
			}
		}

		if err := tx.Commit(); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		name, _ := queries.GetBallotName(ctx, ballotID)
		pages.VotingSuccess(name).Render(ctx, w)

	}))
	http.Handle("/vote/{ballot}", auth.MustBeAuthenticated(db, func(w http.ResponseWriter, r *http.Request) {

		ctx := r.Context()
		ballotID, err := uuid.Parse(r.PathValue("ballot"))
		if err != nil {
			http.Error(w, "invalid ballot id", http.StatusBadRequest)
			return
		}

		status, err := queries.GetBallotStatus(ctx, ballotID)
		if err != nil {
			http.Error(w, "ballot not found", http.StatusNotFound)
			return
		}
		if status != "open" {
			http.Error(w, "ballot is not open", http.StatusForbidden)
			return
		}

		candidates, err := queries.GetBallotCandidates(ctx, ballotID)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		name, _ := queries.GetBallotName(ctx, ballotID)

		pages.VotingPage(name, candidates).Render(r.Context(), w)
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
	http.Handle("/", templ.Handler(pages.Landing()))
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
