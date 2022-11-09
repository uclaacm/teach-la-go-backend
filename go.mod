module github.com/uclaacm/teach-la-go-backend

// +heroku goVersion 1.14
go 1.14

require (
	cloud.google.com/go/firestore v1.2.0
	firebase.google.com/go v3.13.0+incompatible
	github.com/google/uuid v1.3.0
	github.com/heroku/x v0.0.25
	github.com/joho/godotenv v1.3.0
	github.com/labstack/echo/v4 v4.7.2
	github.com/labstack/gommon v0.3.1
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.8.1
	github.com/uclaacm/teach-la-go-backend-tinycrypt v1.0.0
	github.com/urfave/cli/v2 v2.11.0
	golang.org/x/net v0.0.0-20221014081412-f15817d10f9b
	google.golang.org/api v0.103.0
	google.golang.org/grpc v1.50.1
)

replace github.com/joho/godotenv => github.com/x1unix/godotenv v1.3.1-0.20200910042738-acd8c1e858a6
