GO=go
COMMIT_SHA=$(shell git rev-parse HEAD)
# COMMIT_SHA=1
# IMAGE_STAGING=powerwork-staging-mpa
# IMAGE_PRODUCTION=powerwork-mpa
# PROJECT_NAME=powerworkamld
# BUCKET_STAGING=daml-bucket
# BUCKET_PRODUCTION=daml-bucket

IMAGE_STAGING=powerwork-staging-mpa
IMAGE_PRODUCTION=powerwork-mpa-p
PROJECT_NAME=powerwork
BUCKET_STAGING=powerwork-stage-bucket
BUCKET_PRODUCTION=powerwork-bucket


help:
	# -web
	#
	# make dev -- start live reload on port 8000
	# make style -- build file
	# make watch -- build file and watch file
	# make stylep -- build file production
	#
	# make dkp build stylep task and push to docker gcr production
	# make dkstage build stylep task and push to docker gcr staging
	# make updkf up docker-compose file production
	# make production clean build and deploy image to gcs

dev:
	goreload -x vendor -x src --all

style:
	npm run dev

watch:
	npm run watch

stylep:
	npm run production

merge-production:
	git checkout master
	git pull -r
	git push origin HEAD:production --force
	git checkout production
	git pull -r

clean:
	rm -f amlp

build: clean
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO) build -o amlp -ldflags '-w -s' main.go

production: build stylep
	docker build -t $(IMAGE_PRODUCTION):$(COMMIT_SHA) .
	docker tag $(IMAGE_PRODUCTION):$(COMMIT_SHA) gcr.io/$(PROJECT_NAME)/$(IMAGE_PRODUCTION):$(COMMIT_SHA)
	gcloud docker -- push gcr.io/$(PROJECT_NAME)/$(IMAGE_PRODUCTION):$(COMMIT_SHA)
	echo gcr.io/$(PROJECT_NAME)/$(IMAGE_PRODUCTION):$(COMMIT_SHA)

updkf:
	scp -i ~/.ssh/amlp docker-production.yml root@165.22.100.167:/var/app

# scp -i ~/.ssh/menfinhub config/config-stage-server.yaml root@128.199.173.150:/var/app/config
# manual upload
# gsutil -h "Cache-Control:public, max-age=31536000" -m cp -n -c /Users/anthoz/Downloads/Create_post/* gs://menfinhub/static/help

#Server docker compose
#docker-compose -f docker-production.yml up -d (RUN on -d backgroud, -f file name)
#docker-compose -f docker-production.yml down (STOP GO)

#ssh -i ~/.ssh/menfin root@209.97.160.43

#