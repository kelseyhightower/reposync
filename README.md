# reposync

## Build

```
GOOS=linux go build -o function .
```

## Deploy

```
gcloud beta functions deploy reposync \
  --entry-point F \
  --stage-bucket ${PROJECT_ID}-pipeline-functions \
  --trigger-http
```
