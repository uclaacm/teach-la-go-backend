# teach-la-go-backend

Hey there! This is the repo for the Go Backend for the Teach LA editor. If you're a frontend dev and are just looking for documentation of our endpoints, you can find that [right here](https://documenter.getpostman.com/view/10224331/SzYgSFSU?version=latest). If you're a backend dev (or prospective dev) looking to get involved, then read on for info on how to get up and running!

# Quickstart

```sh
go get github.com/uclaacm/teach-la-go-backend

cd $GOPATH/src/github.com/uclaacm/teach-la-go-backend
# add a remote here if you aren't going to use ours

go get -d ./...
go build

# source your credentials or create your own

./teach-la-go-backend
```

For formatting, we use `gofmt`. For linting, we use `golint` and `golangci-lint`.

# Developer Setup

Here's what you need and how to **build** the project:
* [git](https://git-scm.com/)
* [Go](https://golang.org/)

```sh
export TLAPATH=${GOPATH}/src/github.com/uclaacm/teach-la-go-backend

git clone git@github.com:uclaacm/teach-la-go-backend.git $TLAPATH
# alternatively, using HTTPS:
# git clone https://github.com/uclaacm/teach-la-go-backend.git

cd $TLAPATH

# go get dependencies
go get -d ./...
# Note: ./... unrolls the current directory.

# build the server for your platform
make
# ...or build it for all platforms
# make all

# run the server
./bin/tlabe --help
```

If you try running the server at this point (with `./teach-la-go-backend`), the program will crash with a message complaining that a DB client could not be opened. To be precise, it will complain with:

```json
{
    "time": "2020-07-10T02:34:46.7357463-07:00",
    "level": "FATAL",
    "prefix": "echo",
    "file": "server.go",
    "line": "37",
    "message": "no $TLACFG environment variable provided"
}
```

To **run** the project for live development - not just build it - one needs to be able to interact with the TeachLA Firebase through service account credentials (usually a single JSON file). These can be obtained during a TeachLA dev team meeting, or by messaging the #go-backend channel on the TLA Slack.

**You must change the file extension to `.env` so our `.gitignore` will prevent it from being accidentally uploaded to the public repo**. Once you have done so, simply enter the file, surround the json with single quotes (`'`), and prepend `export TLACFG=` to the first file. It should look something like:

```sh
export TLACFG='{
    // ...
}'
```

You can now run the server you built!

## Testing

Development is largely test-driven. Any code you contribute should have tests to go with it. Tests should be placed in another file in the same directory with the naming convention `my_file_name_test.go`.

Run tests with the following commands:

```sh
# run all tests
go test ./...

# do so with **verbosity**
go test -v ./...

# run a specific test
go test -run TestNameHere
```

With this, you can build, test, and run the actual backend. If you'd like to get working, you can stop reading here. Otherwise, you can scan through some of the FAQ below.

## Go FAQ

Go is an new language to a great many people. Hopefully the questions you have might be answered below:

### Q: Why even use Go?

Go is a modern, well-abstracted language for writing performant backends and web applications. It has un*paralleled* support for parallelism out of the box -- so much so that it provides primitive types for concurrency out of the box. It is compiled and garbage-collected. All binaries are statically linked.

### Q: What are the naming conventions?

Go has some interesting naming conventions. Here's the clif notes:
* Filenames should be `snake_cased.go`
* Function and variable names should be `camelCased()`
* Any functions and variables in a package with names that are `UpperCamelCased` are exported types and can be imported to other packages.
* Exported constants should be in `CAPS`.
* Test filenames should be `snake_cased_and_end_in_test.go`
* Test routine names should begin with `Test` and mention the function or feature they intend to test (i.e. `TestCreateProgram`).

### Q: What should my coding style be?

Please, please, **please** use `gofmt` to format your code. Use `golint` (or, better yet, `golangci-lint`) for linting. This makes life easier down the line when others read your code.

Also make sure that you:
* Comment all exported symbols.
* Keep names idiomatic.

### Q: Where should I put my code?

We keep our code for handlers in the `db` folder. Each file name describes the class of handlers and associated database types it deals with. For example, `db/program.go` contains the definition for the `Program` type and all handlers that work with it.

If you have any code that extends functionality of an existing package -- say, `pkg` -- place it in another folder `pkgext`. You can take a look at `httpext` for an example of this.

## Didn't answer your question?

If you're on the TeachLA Slack, feel free to @leo with any questions or shoot a message off to the #go-backend channel.

If you're not on our Slack, feel free to shoot an email off to @krashanoff on GitHub.