[![Coverage Status](https://coveralls.io/repos/github/uclaacm/teach-la-go-backend/badge.svg?branch=master)](https://coveralls.io/github/uclaacm/teach-la-go-backend?branch=master)

# teach-la-go-backend

Hey there! This is the repo for the Go Backend for the Teach LA editor. If you're a frontend dev and are just looking for documentation of our endpoints, you can find that [right here](https://documenter.getpostman.com/view/10224331/TW6xmnn2). If you're a backend dev (or prospective dev) looking to get involved, then read on for info on how to get up and running!

# Quickstart

To run the backend locally, you can download the latest build for your system from the
[releases page](https://github.com/uclaacm/teach-la-go-backend/releases/latest). After
doing so, follow along with the guide below!

```sh
$ # compile the server
$ make
$ # run the server
$ ./bin/tlabe -h
NAME:
   Teach LA Go Backend - tlabe [options]

USAGE:
   tlabe [global options] [arguments...]

VERSION:
   1.0.0

DESCRIPTION:
   Teach LA's editor backend.

GLOBAL OPTIONS:
   --dotenv value, -e value  Specify a path to a dotenv file to specify credentials
   --json value, -j value    Specify a path to a JSON file to specify credentials
   --verbose, -v             Change the log level used by echo's logger middleware (default: false)
   --port value, -p value    Change the port number (default: "8081")
   --help, -h                Show help (default: false)
   --version, -V             Print the version and exit (default: false)

$ ./bin/tlabe -j credentials.json
⇨ http server started on [::]:8081

$ # from here, you can start up the frontend using your own backend!
```

# Developer Setup

Here's what you need and how to **build** the project:
* [git](https://git-scm.com/)
* [go](https://golang.org/)
* [make](https://www.gnu.org/software/make/manual/make.html) (optional)

Here's what you need to get your code PR-ready and **contribute** to the project:
* For formatting, we use `gofmt`. This is included with your installation of [go](https://golang.org/)
* For linting, we use [`golangci-lint`](https://github.com/golangci/golangci-lint).

```sh
git clone https://github.com/uclaacm/teach-la-go-backend.git
cd teach-la-go-backend

# build the server for your platform
make
# ...or build it for all platforms with the below command:
# make all

# run the server
./bin/tlabe --help
```

If you try running the server at this point (with `./bin/tlabe`), the program will crash with a message complaining that a DB client could not be opened. To be precise, it will complain with:

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

Once you have acquired a copy of `credentials.json` or otherwise, you can specify the credentials file location:

```sh
./tlabe -j credentials.json
```

You can now run the server you built!

## Testing

Development is test-focused. Any code you contribute should have tests to go with it. Tests should be placed in another file in the same directory with the naming convention `my_file_name_test.go`.

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

## FAQ

### Q: What are the naming conventions?

Go has some interesting naming conventions. Here's the clif notes:
* Filenames should be `snake_cased.go`
* Function and variable names should be `camelCased()`
* Any functions and variables in a package with names that are `UpperCamelCased` are exported types and can be imported to other packages.
* Exported constants should be in `CAPS`.
* Test filenames should be `snake_cased_and_end_in_test.go`
* Test routine names should begin with `Test` and mention the function or feature they intend to test (i.e. `TestCreateProgram`).

### Q: What should my coding style be?

Please, please, **please** use `gofmt` to format your code. Use `golangci-lint` for linting. This makes life easier down the line when others read your code.

Make sure that you:
* Comment all exported symbols.
* Keep names idiomatic.

### Q: Where should I put my code?

We keep our code for handlers in the `db` folder. Each file name describes the class of handlers and associated database types it deals with. For example, `db/program.go` contains the definition for the `Program` type and all handlers that work with it.

If you have any code that extends functionality of an existing package -- say, `pkg` -- place it in another folder `pkgext`. You can take a look at `httpext` for an example of this.
