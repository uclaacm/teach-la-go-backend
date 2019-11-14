# teach-la-go-backend

Hey there! This is the repo for our **experimental** Go Backend, which we're using for our online editor. Eventually, the goal of this project is to replace [the current Express-based backend](https://github.com/uclaacm/TeachLAJSBackend), bringing it up to feature parity and using all the benefits that Go provides!

## Developer Setup

Requirements:
* [git](https://git-scm.com/)
* [Go](https://golang.org/)

To get coding:

```sh
# clone the repo
git clone git@github.com:uclaacm/teach-la-go-backend.git
cd teach-la-go-backend

# go get dependencies
go get -d ./...

# build and run the server
go build
./teach-la-go-backend
```