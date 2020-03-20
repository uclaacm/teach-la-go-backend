# teach-la-go-backend

Hey there! This is the repo for our Go Backend, which we're using for our online editor.

If you're on the TeachLA Slack, feel free to pop into the #go-backend channel with any and all questions. Thanks!

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
If you try running the server at this point (with `./teach-la-go-backend`), you'll probably get a message like this: `...no $PORT environment variable provided.`. To **run** the project, one needs to set the port to run the backend on and have the ability to interact with the TeachLA Firebase through service account credentials, which are provided via the `$TLACFG` environment variable. These can be obtained during a TeachLA dev team meeting, or by messaging the #go-backend channel on the TLA Slack.

Once acquired, set the variables:

```
export PORT=8081
export TLACFG='my secret stuff'
```

**It is recommended that you put these commands in a `*.env` file to avoid having to run these manually.** To set your environment variables through the file, simply run `source MYFILENAME.env`.

You can now run the server you built!

## Testing

Run the following command to run tests:

```sh
# run all tests
go test ./...

# run a specific test
go test ./server_test.go

# run tests with verbose output
go test -v ./server_test.go
```

## Documentation

For a formal description of the endpoints for our backend, you can scan through the documentation provided in the repository's website link on GitHub.