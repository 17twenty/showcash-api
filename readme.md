# showcash-api

## Getting Started

```bash
go get ./...
go build ./cmd/...
./showcash-api
```

## Issues

```bash
curl -H "Authorization: Bearer yJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6Im5pY2tAc2hvd2Nhc2guaW8iLCJVc2VySUQiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAiLCJVc2VyU3RhdHVzIjoxLCJleHAiOjE1NzgyOTAxNzQsImlhdCI6MTU3ODI4OTI3NH0.G5qVCq2-ZGNHKdmWi70Nvncigqkjxi-8GWrHsAnDhk3OmLrCCVMgYHqIgYctqFTMuAID9vgj4F0PX571SNmKeg" localhost:8080/api/me  --cookie "jwt-token=eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJFbWFpbCI6Im5pY2tAc2hvd2Nhc2guaW8iLCJVc2VySUQiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAiLCJVc2VyU3RhdHVzIjoxLCJleHAiOjE1NzgyOTAxNzQsImlhdCI6MTU3ODI4OTI3NH0.G5qVCq2-ZGNHKdmWi70Nvncigqkjxi-8GWrHsAnDhk3OmLrCCVMgYHqIgYctqFTMuAID9vgj4F0PX571SNmKeg; Path=/; Expires=Wed, 05 Feb 2020 05:41:14 GMT"
```

Currently i borrowed messy code from Quicka... this code requires sending the authorisation header AND a cookie value... this is obviously
dumb as fuck so need to fix that at some point from perspective of sucking less for Vue.