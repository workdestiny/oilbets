package app

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/workdestiny/oilbets/entity"
	"github.com/workdestiny/oilbets/service"

	"github.com/acoshift/header"
	"github.com/acoshift/middleware"
	"github.com/moonrhythm/hime"
	"github.com/workdestiny/oilbets/repository"
	"golang.org/x/net/xsrftoken"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func logHTTP(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		timeStart := time.Now().In(loc)
		lrw := newLoggingResponseWriter(w)
		h.ServeHTTP(lrw, r)
		if strings.HasPrefix(r.URL.Path, "/-") {
			return
		}
		timeEnd := time.Now().In(loc)
		clientIP, port, _ := net.SplitHostPort(r.RemoteAddr)
		statusCode := lrw.statusCode
		request := r.RequestURI
		fmt.Printf("%s | %v | %v | %3d | %13v | %s | %s\n", timeEnd.Format("2006-01-02 15:04:05"), clientIP, port, statusCode, timeEnd.Sub(timeStart), r.Method, request)
	})
}

func noCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}

func securityHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(header.XFrameOptions, "deny")
		w.Header().Set(header.XXSSProtection, "1; mode=block")
		w.Header().Set(header.XContentTypeOptions, "nosniff")
		h.ServeHTTP(w, r)
	})
}

func methodFilter(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodOptions:
			h.ServeHTTP(w, r)
		default:
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})
}

func cacheHeader(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(header.CacheControl, "public, max-age=31536000")
		h.ServeHTTP(w, r)
	})
}

func fileHandler(name string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, name)
	})
}

func panicRecovery(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Println(err)
				debug.PrintStack()

				me := getUser(r.Context())
				// send massage to Discord Channel Chat
				service.SendErrorToDiscord(me, fmt.Sprintf("%s || %v", r.RequestURI, err))

				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("X-Content-Type-Options", "nosniff")
				w.Header().Set("Cache-Control", "private, no-cache, no-store, max-age=0")
				w.WriteHeader(http.StatusInternalServerError)
				io.Copy(w, bytes.NewReader(errorPage))
			}
		}()
		h.ServeHTTP(w, r)
	})
}

func getCookie(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userID := getUserID(ctx)
		nctx := WithMyID(ctx, userID)

		vsID := getVisitorID(ctx)
		if vsID == "" {
			vsID = uuid.New().String()
			SaveSessionVisitorID(ctx, vsID)
		}

		nctx = WithVisitorID(nctx, vsID)
		h.ServeHTTP(w, r.WithContext(nctx))
	})
}

func xsrf(baseURL, xsrfSecret string) middleware.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := GetMyID(r.Context())
			if r.Method == http.MethodPost {
				origin := r.Header.Get(header.Origin)
				if len(origin) > 0 {
					if origin != baseURL {
						http.Error(w, "Not allow cross-site post", http.StatusBadRequest)
						return
					}
				}
				x := r.FormValue("X")
				if !xsrftoken.Valid(x, xsrfSecret, id, r.URL.Path) {
					http.Error(w, "invalid xsrf token, go back, refresh and try again...", http.StatusBadRequest)
					return
				}
				h.ServeHTTP(w, r)
				return
			}
			token := xsrftoken.Generate(xsrfSecret, id, r.URL.Path)
			ctx := WithTokenXSRF(r.Context(), token)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func fetchUser(db *sql.DB) middleware.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			rCtx := r.Context()
			me := &entity.Me{}

			id := GetMyID(rCtx)
			if id != "" {
				me = repository.GetUser(db, id)
				if me.ID == "" {
					removeSession(r.Context())
				}

			}

			ctx := WithUser(rCtx, me)
			ctx = WithUserRole(ctx, entity.Role(me.Role))
			ctx = WithUserFollowTopic(ctx, false)

			if me.Count.Topic > 0 {
				ctx = WithUserFollowTopic(ctx, true)
			}

			h.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

func userAgent(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Header.Get("user-agent")

		if value == "" {
			value = "nil"
		}

		nctx := WithUserAgent(r.Context(), value)
		h.ServeHTTP(w, r.WithContext(nctx))
	})
}

func isPublic(h http.Handler) http.Handler {
	return hime.Handler(func(ctx *hime.Context) error {
		if GetMyID(ctx) != "" {
			return ctx.RedirectTo("discover")
		}
		return ctx.Handle(h)
	})
}

func isUser(h http.Handler) http.Handler {
	return hime.Handler(func(ctx *hime.Context) error {
		if GetMyID(ctx) == "" {
			return ctx.RedirectTo("signin")
		}
		return ctx.Handle(h)
	})
}

func isAdmin(h http.Handler) http.Handler {
	return hime.Handler(func(ctx *hime.Context) error {
		role := GetUserRole(ctx)
		if role != entity.RoleAdmin && role != entity.RoleAgent {
			return ctx.RedirectTo("notfound")
		}
		return ctx.Handle(h)
	})
}

func isAnonymous(h http.Handler) http.Handler {
	return hime.Handler(func(ctx *hime.Context) error {
		if GetMyID(ctx) != "" {
			return ctx.RedirectTo("discover")
		}
		return ctx.Handle(h)
	})
}

func isBookbank(h http.Handler) http.Handler {
	return hime.Handler(func(ctx *hime.Context) error {
		if !getUser(ctx).IsVerifyBookBank {
			return ctx.RedirectTo("discover")
		}
		return ctx.Handle(h)
	})
}

// DefaultCacheControl sets default cache-control header
func DefaultCacheControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(header.CacheControl, "no-cache, no-store, must-revalidate")
		h.ServeHTTP(w, r)
	})
}
