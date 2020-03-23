# showcash-api

## Getting Started

```bash
go get ./...
go build ./cmd/...
./showcash-api
2020/03/20 13:57:07 Starting Showcash API
2020/03/20 13:57:07.898245 core.go:63: Showcashing it on port 8080...
2020/03/20 13:57:11.447623 core.go:68: Got data
2020/03/20 13:57:11.459790 core.go:78: Uploaded... Overview.png
```

## Deployment

```bash
aws --profile showcash \
ecr get-login-password \
  --region ap-southeast-2 | docker login \
  --username AWS \
  --password-stdin 386569642910.dkr.ecr.ap-southeast-2.amazonaws.com
make publish
```

## Issues
