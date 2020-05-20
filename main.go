package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"cloud.google.com/go/storage"
	"github.com/acoshift/configfile"
	"github.com/acoshift/csrf"
	"github.com/acoshift/ds"
	"github.com/acoshift/probehandler"
	"github.com/go-redis/redis"
	_ "github.com/lib/pq"
	"github.com/moonrhythm/hime"
	"github.com/moonrhythm/session"
	redisstore "github.com/moonrhythm/session/store/goredis"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/workdestiny/oilbets/app"
	"github.com/workdestiny/oilbets/service"
)

// set time zone
func main() {
	loc, _ := time.LoadLocation("Asia/Bangkok")

	configValue := configfile.NewYAMLReader("config/config-application.yaml")

	ctx := context.Background()

	// get GCloud
	googleConfig, err := google.JWTConfigFromJSON(getFileConfigSecret().Bytes(configValue.String("service_account")), datastore.ScopeDatastore, storage.ScopeReadWrite)
	if err != nil {
		log.Fatal(err)
	}

	tokenSource := googleConfig.TokenSource(ctx)

	client, err := ds.NewClient(ctx, configValue.String("datastorne_name"), option.WithTokenSource(tokenSource))
	if err != nil {
		log.Fatal(err)
	}

	storageClient, err := storage.NewClient(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		log.Fatal(err)
	}

	bucket := storageClient.Bucket(configValue.String("bucket_name"))
	bucketSecret := storageClient.Bucket(configValue.String("bucket_name_secret"))

	sessionHost := configValue.String("session_host")
	redisPrefix := configValue.String("session_prefix")
	redisClient := redis.NewClient(&redis.Options{
		Addr:       sessionHost,
		MaxRetries: 3,
		PoolSize:   6,
	})
	defer redisClient.Close()

	db, err := sql.Open("postgres", configValue.String("db"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(0)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.Ping()

	log.Println(db)

	app.RunRedisX(redisClient, db)

	appHime := hime.New()

	appFactory := &app.App{
		Location:     loc,
		Domain:       configValue.String("domain"),
		BaseURL:      configValue.String("base_url"),
		ShortenerURL: configValue.String("shortener_url"),
		SQL:          db,
		Redis:        redisClient,
		RedisPrefix:  redisPrefix,
		Ds:           client,
		SessionConfig: session.Config{
			Store: redisstore.New(redisstore.Config{
				Client: redisClient,
				Prefix: redisPrefix,
			}),
			HTTPOnly: true,
			Secure:   session.PreferSecure,
			Proxy:    true,
			MaxAge:   60 * 24 * time.Hour,
			Path:     "/",
			Rolling:  true,
			Keys:     [][]byte{configValue.Bytes("session_key")},
			Secret:   configValue.Bytes("session_secret"),
			SameSite: http.SameSiteLaxMode,
		},
		Bucket: app.Bucket{
			Storage: bucket,
			Name:    configValue.String("bucket_name"),
		},
		BucketSecret: app.Bucket{
			Storage: bucketSecret,
			Name:    configValue.String("bucket_name_secret"),
		},
		CSRFConfig: csrf.Config{
			Origins: []string{
				configValue.String("domain"),
				"mahaporn.com",
				"mahaporn.local:8080",
			},
			IgnoreProto: true,
		},
		FacebookAppID:      configValue.String("facebook_appid"),
		FacebookSecret:     configValue.String("facebook_secret"),
		GoogleClient:       configValue.String("google_client"),
		GoogleSecret:       configValue.String("google_secret"),
		GoogleCookieSecret: configValue.Bytes("google_cookie_secret"),
		ErrorPage:          service.MustLoadBytesFromFile("template/error.html"),
		Hime:               appHime,
		Static:             static("public/mix-manifest.json"),
	}

	service.New(&service.Config{
		Location: loc,
	})

	appHime.Template().
		Funcs(appFactory.TemplateFuncs()).
		ParseConfigFile("settings/web/template.yaml")

	appHime.
		ParseConfigFile("settings/web/routes.yaml").
		ParseConfigFile("settings/web/server.yaml").
		Handler(appFactory.Handler())

	probe := probehandler.New()
	health := http.NewServeMux()
	health.Handle("/", probehandler.Success())
	health.Handle("/readiness", probe)
	go http.ListenAndServe(":18080", health)

	appHime.
		GracefulShutdown().
		Notify(probe.Fail)

	err = appHime.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}

func getFileConfigSecret() *configfile.Reader {
	return configfile.NewReader("config/secret")
}

func static(filename string) map[string]string {
	s := make(map[string]string)
	bs, _ := ioutil.ReadFile(filename)
	json.Unmarshal(bs, &s)
	return s
}
