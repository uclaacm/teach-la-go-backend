module github.com/uclaacm/teach-la-go-backend

// +heroku goVersion 1.14
go 1.14

require (
	cloud.google.com/go/firestore v1.2.0
	cloud.google.com/go/iam v0.3.0 // indirect
	firebase.google.com/go v3.13.0+incompatible
	github.com/google/uuid v1.1.2
	github.com/heroku/x v0.0.25
	github.com/joho/godotenv v1.3.0
	github.com/labstack/echo/v4 v4.1.16
	github.com/labstack/gommon v0.3.0
	github.com/mattn/go-colorable v0.1.7 // indirect
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	github.com/uclaacm/teach-la-go-backend-tinycrypt v1.0.0
	github.com/urfave/cli/v2 v2.3.0
	github.com/valyala/fasttemplate v1.2.0 // indirect
	golang.org/x/crypto v0.0.0-20200709230013-948cd5f35899 // indirect
	golang.org/x/net v0.0.0-20220607020251-c690dde0001d
	google.golang.org/api v0.84.0
	google.golang.org/grpc v1.47.0
)

replace github.com/joho/godotenv => github.com/x1unix/godotenv v1.3.1-0.20200910042738-acd8c1e858a6
