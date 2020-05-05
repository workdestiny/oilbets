package app

import (
	"database/sql"
	"time"

	"github.com/moonrhythm/hime"

	"github.com/acoshift/csrf"

	"cloud.google.com/go/storage"
	"github.com/acoshift/ds"
	"github.com/go-redis/redis"
	"github.com/moonrhythm/session"
)

// App is struct config for app
type App struct {
	Location           *time.Location
	SessionConfig      session.Config
	Domain             string
	BaseURL            string
	ShortenerURL       string
	SQL                *sql.DB
	Redis              *redis.Client
	RedisPrefix        string
	Ds                 *ds.Client
	Bucket             Bucket
	BucketSecret       Bucket
	CSRFConfig         csrf.Config
	FacebookAppID      string
	FacebookSecret     string
	GoogleClient       string
	GoogleSecret       string
	GoogleCookieSecret []byte
	ErrorPage          []byte
	Hime               *hime.App
	Static             map[string]string
}

// Bucket is GCloud Storage and Name Storage
type Bucket struct {
	Storage *storage.BucketHandle
	Name    string
}

var (
	loc                *time.Location
	baseURL            string
	shortenerURL       string
	domain             string
	db                 *sql.DB
	myRedis            *redis.Client
	redisPrefix        string
	client             *ds.Client
	bucket             Bucket
	bucketSecret       Bucket
	csrfConfig         csrf.Config
	facebookAppID      string
	facebookSecret     string
	googleClient       string
	googleSecret       string
	googleCookieSecret []byte
	errorPage          []byte
)

func initConfig(c *App) {
	loc = c.Location
	baseURL = c.BaseURL
	shortenerURL = c.ShortenerURL
	domain = c.Domain
	client = c.Ds
	myRedis = c.Redis
	redisPrefix = c.RedisPrefix
	db = c.SQL
	bucket = c.Bucket
	bucketSecret = c.BucketSecret
	csrfConfig = c.CSRFConfig
	facebookAppID = c.FacebookAppID
	facebookSecret = c.FacebookSecret
	googleClient = c.GoogleClient
	googleSecret = c.GoogleSecret
	googleCookieSecret = c.GoogleCookieSecret
	errorPage = c.ErrorPage

}
