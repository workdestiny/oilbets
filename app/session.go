package app

import (
	"context"

	"github.com/moonrhythm/session"
)

var (
	// SessionName is o_sess
	SessionName = "o_sess"
)

func getUserID(ctx context.Context) string {
	return getSession(ctx).GetString("userid")
}

func getReferrer(ctx context.Context) string {
	return getSession(ctx).GetString("referrer")
}

func getVisitorID(ctx context.Context) string {
	return getSession(ctx).GetString("vsid")
}

func getSession(ctx context.Context) *session.Session {
	s, err := session.Get(ctx, SessionName)
	must(err)
	return s
}

// RemoveSession Remove Session
func removeSession(ctx context.Context) *session.Session {
	sess := getSession(ctx)
	sess.Del("userid")
	return sess
}

// RemoveSession Remove Session
func removeSessionReferrer(ctx context.Context) *session.Session {
	sess := getSession(ctx)
	sess.Del("referrer")
	return sess
}

// SaveSession is set userid in session
func SaveSession(ctx context.Context, userID string) {
	sess := getSession(ctx)
	sess.Set("userid", userID)
}

// SaveSessionReferrer is set referrer in session
func SaveSessionReferrer(ctx context.Context, referrer string) {
	sess := getSession(ctx)
	sess.Set("referrer", referrer)
}

// SaveSessionVisitorID is set userid in session
func SaveSessionVisitorID(ctx context.Context, code string) {
	sess := getSession(ctx)
	sess.Set("vsid", code)
}

func addSuccess(f *session.Flash, s string) {
	f.Clear()
	f.Add("Success", s)
}
