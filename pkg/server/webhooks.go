package server

import (
	"fmt"
	"html"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/lithammer/shortuuid/v4"
	"github.com/rs/zerolog/log"
	"github.com/urfave/cli/v3"
	"google.golang.org/grpc/credentials"

	"github.com/tzrikka/thrippy/pkg/client"
	"github.com/tzrikka/thrippy/pkg/links/github"
)

const (
	timeout = time.Second * 3
)

type httpServer struct {
	httpPort int // To initialize the HTTP server.

	grpcAddr  string // To communicate with the secrets manager.
	grpcCreds credentials.TransportCredentials

	redirectURL string // The server's OAuth callback URL.
	fallbackURL string // Optional destination for OAuth callbacks without a state.
}

func newHTTPServer(cmd *cli.Command) *httpServer {
	return &httpServer{
		httpPort: cmd.Int("webhook-port"),

		grpcAddr:  cmd.String("grpc-addr"),
		grpcCreds: client.GRPCCreds(cmd),

		redirectURL: fmt.Sprintf("https://%s/callback", cmd.String("webhook-addr")),
		fallbackURL: cmd.String("fallback-url"),
	}
}

// run starts an HTTP server for OAuth webhooks.
// This is blocking, to keep the Thrippy server running.
func (s *httpServer) run() error {
	http.HandleFunc("GET /callback", s.oauthExchangeHandler)
	http.HandleFunc("GET /start", s.oauthStartHandler)
	http.HandleFunc("POST /start", s.oauthStartHandler)

	server := &http.Server{
		Addr:         net.JoinHostPort("", strconv.Itoa(s.httpPort)),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
	}

	log.Info().Msgf("HTTP server listening on port %d", s.httpPort)
	log.Info().Msgf("OAuth callback URL: %s", s.redirectURL)
	err := server.ListenAndServe()
	if err != nil {
		log.Err(err).Send()
	}

	return err
}

// oauthStartHandler starts a 3-legged OAuth 2.0 flow by redirecting the client
// to the authorization endpoint of a third-party service. The incoming request's
// method may be GET or POST, but the resulting redirection should always be GET.
func (s *httpServer) oauthStartHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	l := log.With().Str("http_method", r.Method).Str("url_path", r.URL.EscapedPath()).Logger()
	l.Info().Msg("received HTTP request")

	// Get the OAuth config corresponding to the link ID in the request's query or body.
	if err := r.ParseForm(); err != nil {
		l.Warn().Err(err).Msg("bad request: form parsing error")
		htmlResponse(w, http.StatusBadRequest, "Form parsing error")
		return
	}

	id := r.FormValue("id")
	if id == "" {
		l.Warn().Msg("bad request: missing ID parameter")
		htmlResponse(w, http.StatusBadRequest, "Missing ID parameter")
		return
	}

	l = l.With().Str("id", id).Logger()
	if _, err := shortuuid.DefaultEncoder.Decode(id); err != nil {
		l.Warn().Err(err).Msg("bad request: ID is an invalid short UUID")
		htmlResponse(w, http.StatusBadRequest, "Invalid ID parameter")
		return
	}

	ctx := l.WithContext(r.Context())
	o, err := client.LinkOAuthConfig(ctx, s.grpcAddr, s.grpcCreds, id)
	if err != nil {
		htmlResponse(w, http.StatusInternalServerError, "&nbsp;")
		return
	}
	if o == nil {
		l.Warn().Err(err).Msg("bad request: link not found")
		htmlResponse(w, http.StatusBadRequest, "Link not found")
		return
	}

	// Redirect based on the OAuth config, using its ID as the state parameter,
	// with an optional (short, opaque, but not secret) memo from the caller.
	o.Config.RedirectURL = s.redirectURL
	state := constructStateParam(id, r.FormValue("memo"))
	http.Redirect(w, r, o.AuthCodeURL(state), http.StatusFound)
	l.Debug().Str("url", o.Config.Endpoint.AuthURL).Msg("redirected HTTP request")
}

// oauthExchangeHandler receives a redirect back from a third-party service's
// authorization endpoint (the 2nd lef of the OAuth 2.0 flow), and exchanges
// the received authorization code for an new access token (the 3rd leg).
func (s *httpServer) oauthExchangeHandler(w http.ResponseWriter, r *http.Request) {
	l := log.With().Str("http_method", r.Method).Str("url_path", r.URL.EscapedPath()).Logger()
	l.Info().Msg("received HTTP request")

	if err := r.ParseForm(); err != nil {
		l.Warn().Err(err).Msg("bad request: form parsing error")
		htmlResponse(w, http.StatusBadRequest, "Form parsing error")
		return
	}

	// First, check for errors reported by the third-party,
	// e.g. the user failed/refused to authorize Thrippy.
	errParam := r.FormValue("error_description")
	if errParam == "" {
		errParam = r.FormValue("error")
	}
	if errParam != "" {
		errParam = html.EscapeString(errParam)
		l.Warn().Msgf("OAuth error: %s", errParam)
		htmlResponse(w, http.StatusBadRequest, errParam)
		return
	}

	// If the state parameter is missing, it means the OAuth flow was not
	// initiated by Thrippy, so we can't do anything with the results.
	state := r.FormValue("state")
	if state == "" {
		l.Warn().Msg("forbidden: missing OAuth state parameter")
		if s.fallbackURL != "" {
			l.Debug().Str("url", s.fallbackURL).Msg("redirected HTTP request")
			http.Redirect(w, r, s.fallbackURL, http.StatusFound)
			return
		}
		htmlResponse(w, http.StatusForbidden, "Missing OAuth state parameter")
		return
	}

	// Parse the state parameter.
	id, memo, err := parseStateParam(state)
	l = l.With().Str("state", id).Str("id", id).Logger()
	if memo != "" {
		l = l.With().Str("memo", memo).Logger()
	}
	if err != nil {
		l.Warn().Err(err).Msg("bad request: state is an invalid short UUID")
		htmlResponse(w, http.StatusBadRequest, "Invalid state parameter")
		return
	}

	ctx := l.WithContext(r.Context())
	o, err := client.LinkOAuthConfig(ctx, s.grpcAddr, s.grpcCreds, id)
	if err != nil {
		htmlResponse(w, http.StatusInternalServerError, "&nbsp;")
		return
	}
	if o == nil {
		l.Warn().Err(err).Msg("bad request: link not found")
		htmlResponse(w, http.StatusBadRequest, "Link not found")
		return
	}

	// Special case: requests to install GitHub apps by users who are
	// not authorized to approve them can't continue. For more details, see:
	// https://docs.github.com/en/apps/using-github-apps/installing-a-github-app-from-a-third-party#requirements-to-install-a-github-app
	setupAction := r.FormValue("setup_action")
	if setupAction == "request" {
		l.Warn().Err(err).Msg("GitHub app installation requested by user who can't approve it")
		htmlResponse(w, http.StatusForbidden, "Installation must be approved by an organization owner")
	}

	// Special case: GitHub apps that use generated JWTs don't require a
	// user or app-installation token (the 3rd leg of the OAuth 2.0 flow).
	installID := r.FormValue("installation_id")
	if (setupAction == "install" || setupAction == "update") && installID != "" {
		l = l.With().Str("setup_action", setupAction).Str("install_id", installID).Logger()
		l.Debug().Msg("successful GitHub app installation")

		// Check the app installation, extract metadata with and about it, and save them.
		ctx := l.WithContext(r.Context())
		url := github.APIBaseURL(github.AuthBaseURL(o))
		if err := client.AddGitHubCreds(ctx, s.grpcAddr, s.grpcCreds, id, installID, url); err != nil {
			htmlResponse(w, http.StatusInternalServerError, "&nbsp;")
			return
		}

		l.Debug().Msg("checked and saved the GitHub installation")
		htmlResponse(w, http.StatusOK, "You may now close this browser tab")
		return
	}

	// Exchange the received authorization code for an access token, potentially
	// including a refresh token (this is the 3rd leg of the OAuth 2.0 flow).
	code := r.FormValue("code")
	if code == "" {
		l.Warn().Any("query", r.URL.Query()).Msg("forbidden: missing OAuth code parameter")
		htmlResponse(w, http.StatusForbidden, "Missing OAuth code parameter")
		return
	}

	o.Config.RedirectURL = s.redirectURL
	token, err := o.Exchange(ctx, code)
	if err != nil {
		l.Warn().Err(err).Msg("OAuth code exchange error")
		htmlResponse(w, http.StatusForbidden, "OAuth code exchange error")
		return
	}
	l.Debug().Msg("successful OAuth token exchange")

	// Check the token, extract metadata with and about it, and save them.
	if err := client.SetOAuthCreds(ctx, s.grpcAddr, s.grpcCreds, id, token); err != nil {
		htmlResponse(w, http.StatusInternalServerError, "&nbsp;")
		return
	}

	l.Debug().Msg("checked and saved OAuth token")
	htmlResponse(w, http.StatusOK, "You may now close this browser tab")
}

func htmlResponse(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)

	title := "Success"
	header := "Success!"
	if status >= http.StatusBadRequest {
		title = "Error"
		header = fmt.Sprintf("%d %s", status, http.StatusText(status))
	}

	if !strings.HasSuffix(msg, ".") {
		msg += "."
	}

	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
		<html>
		<head>
			<title>%s</title>
		</head>
		<body>
			<h1>%s</h1>
			<p>%s</p>
		</body>
		</html>`,
		title, header, msg)
}

func constructStateParam(id, memo string) string {
	state := id
	if memo != "" {
		state += "_" + memo
	}
	return state
}

func parseStateParam(state string) (id, memo string, err error) {
	s := strings.SplitN(state, "_", 2)
	if len(s) == 1 {
		s = append(s, "")
	}

	id, memo = s[0], s[1]
	_, err = shortuuid.DefaultEncoder.Decode(id)
	return
}
