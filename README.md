# teach-la-go-backend

Hey there! This is the repo for our **experimental** Go Backend, which we're using for our online editor. Eventually, the goal of this project is to replace [the current Express-based backend](https://github.com/uclaacm/TeachLAJSBackend), bringing it up to feature parity and using all the benefits that Go provides!

If you're on the TeachLA Slack, feel free to @leo with any questions. Thanks!

# Developer Setup

Here's what you need and how to **build** the project:
* [git](https://git-scm.com/)
* [Go](https://golang.org/)

```sh
go get github.com/uclaacm/teach-la-go-backend
# alternatively, using git:
# git clone git@github.com:uclaacm/teach-la-go-backend.git
# OR
# git clone https://github.com/uclaacm/teach-la-go-backend.git

cd $GOPATH/src/github.com/uclaacm/teach-la-go-backend

# set up git pre-commit hook
chmod +x hooks/pre-commit
cp hooks/pre-commit .git/hooks/

# go get dependencies
go get -d ./...
# Note: ./... unrolls the current directory.

# build the server
go build

# run the server
./teach-la-go-backend

```
If you try running the server at this point (with `./teach-la-go-backend`), you'll probably get a message like this: `could not find firebase config file! Did you set your CFGPATH variable? stat : no such file or directory`. To **run** the project, one needs to be able to interact with the TeachLA Firebase through service account credentials (usually a single JSON file). These can be obtained during a TeachLA dev team meeting, or by messaging the #go-backend channel on the TLA Slack. 

Once acquired, save the JSON file in the root directory. **It is recommended that you chnage the file extension to `.env` so `gitignore` will prevent it from being accidentally uploaded to the public repo**. Once you have done that, specify the location of your credentials by setting the environment variable `$CFGPATH`:

```
export CFGPATH=/path/to/creds.json
```

You can now run the server you built!

## Testing

Run the following command to run tests:

```sh
# run all tests
go test ./...

# run a specific test
go test ./server_test.go

# run tests with log output
go test -v ./server_test.go
```

With this, you can build, test, and run the actual backend. If you'd like to get working, you can stop reading here. Otherwise, you can scan through the documentation provided in the repository's website link on GitHub.