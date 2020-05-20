package app

import (
	"context"

	"github.com/acoshift/ds"
	"github.com/workdestiny/oilbets/entity"
)

type contextKey int

const (
	ckMyID contextKey = iota
	ckUserAgent
	ckTokenXSRF
	ckDs
	ckUser
	ckRole
	ckFollowTopic
	ckVisitorID
)

// WithMyID id put to Context
func WithMyID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ckMyID, id)
}

// WithUserRole role put to Context
func WithUserRole(ctx context.Context, role entity.Role) context.Context {
	return context.WithValue(ctx, ckRole, role)
}

// WithUserAgent userAgent put to Context
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, ckUserAgent, userAgent)
}

// WithUserFollowTopic follow topic put to Context
func WithUserFollowTopic(ctx context.Context, isFollow bool) context.Context {
	return context.WithValue(ctx, ckFollowTopic, isFollow)
}

// WithTokenXSRF token put to Context
func WithTokenXSRF(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, ckTokenXSRF, token)
}

// NewDatastoreContext put ds to Context
func NewDatastoreContext(ctx context.Context, ds *ds.Client) context.Context {
	return context.WithValue(ctx, ckDs, ds)
}

// WithUser put user to Context
func WithUser(ctx context.Context, me *entity.Me) context.Context {
	return context.WithValue(ctx, ckUser, me)
}

// WithVisitorID put user to Context
func WithVisitorID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ckVisitorID, id)
}

// GetMyID เอา ID user ออกจาก Context เพื่อเอาไปใช้
func GetMyID(ctx context.Context) string {
	return ctx.Value(ckMyID).(string)
}

// GetUserRole เอา Role user ออกจาก Context เพื่อเอาไปใช้
func GetUserRole(ctx context.Context) entity.Role {
	return ctx.Value(ckRole).(entity.Role)
}

// GetUserAgent Get userAgent in Context
func GetUserAgent(ctx context.Context) string {
	return ctx.Value(ckUserAgent).(string)
}

// GetUserFollowTopic Get isFollow topic in Context
func GetUserFollowTopic(ctx context.Context) bool {
	return ctx.Value(ckFollowTopic).(bool)
}

// GetTokenXSRF Get token in Context
func getTokenXSRF(ctx context.Context) string {
	return ctx.Value(ckTokenXSRF).(string)
}

// GetDatastore Get ds in Context
func GetDatastore(ctx context.Context) *ds.Client {
	return ctx.Value(ckDs).(*ds.Client)
}

func getUser(ctx context.Context) *entity.Me {
	return ctx.Value(ckUser).(*entity.Me)
}

//GetVisitorID Get in Context
func GetVisitorID(ctx context.Context) string {
	return ctx.Value(ckVisitorID).(string)
}
