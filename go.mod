module github.com/uclaacm/teach-la-go-backend

// +heroku goVersion 1.14
go 1.14

require (
	cloud.google.com/go v0.61.0 // indirect
	cloud.google.com/go/firestore v1.2.0
	firebase.google.com/go v3.13.0+incompatible
	github.com/google/uuid v1.1.1
	github.com/heroku/x v0.0.25
	github.com/joho/godotenv v1.3.0
	github.com/labstack/echo/v4 v4.1.16
	github.com/labstack/gommon v0.3.0
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.5.1
	github.com/uclaacm/teach-la-go-backend-tinycrypt v1.0.0
	github.com/urfave/cli/v2 v2.11.0
	github.com/valyala/fasttemplate v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/net v0.0.0-20200707034311-ab3426394381
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
	golang.org/x/tools v0.0.0-20200725200936-102e7d357031 // indirect
	google.golang.org/api v0.29.0
	google.golang.org/genproto v0.0.0-20200724131911-43cab4749ae7 // indirect
	google.golang.org/grpc v1.30.0
	gopkg.in/yaml.v2 v2.2.3 // indirect
)

replace github.com/joho/godotenv => github.com/x1unix/godotenv v1.3.1-0.20200910042738-acd8c1e858a6
