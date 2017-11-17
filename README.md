# reposync

## Build

```
GOOS=linux go build -o function .
```

## Package


```
mkdir reposync-cloud-function-0.0.1
```

```
mv function reposync-cloud-function-0.0.1/
cp index.js reposync-cloud-function-0.0.1/
```

```
zip -r -9 reposync-cloud-function-0.0.1.zip reposync-cloud-function-0.0.1/
```

## Deploy

```
wget https://github.com/kelseyhightower/reposync/releases/download/0.0.1/reposync-cloud-function-0.0.1.zip 
```

```
unzip reposync-cloud-function-0.0.1.zip
```

```
cd reposync-cloud-function-0.0.1
```

```
gcloud beta functions deploy reposync \
  --entry-point F \
  --stage-bucket ${PROJECT_ID}-pipeline-functions \
  --trigger-http
```
