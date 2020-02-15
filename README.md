# teach-la-go-backend

Hey there! This is the repo for our **experimental** Go Backend, which we're using for our online editor. Eventually, the goal of this project is to replace [the current Express-based backend](https://github.com/uclaacm/TeachLAJSBackend), bringing it up to feature parity and using all the benefits that Go provides!

If you're on the TeachLA Slack, feel free to @leo with any questions. Thanks!

# Developer Setup

Here's what you need and how to **build** the project:
* [git](https://git-scm.com/)
* [Go](https://golang.org/)

```sh
git clone git@github.com:uclaacm/teach-la-go-backend.git
# alternatively, using HTTPS:
# git clone https://github.com/uclaacm/teach-la-go-backend.git

cd teach-la-go-backend

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
export CFGPATH=./secret.env
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

With this, you can build, test, and run the actual backend. If you'd like to get working, you can stop reading here. Otherwise, you can scan through the documentation below.

# About the backend

Response codes we use:

Event | `net/http` Constant | Response Code
---|---|:-:
Nominal | `http.StatusOK` | `200`
Bad request | `http.StatusBadRequest` | `400`
Missing resource | `http.StatusNotFound` | `404`
Something unexpected happened server side | `http.StatusInternalServerError` | `500`

## What files do what
* Descriptions of document types can be found in `lib/user.go` and `lib/program.go`.
* Endpoint functionality is divided up into `lib/userManagement.go` and `lib/programManagement.go`.
* All tests are of the form `fileToTestName_test.go`.
* Middleware that handles certain request-specific options prior to handing off the request to the default handler can be found in `middleware/`.
* `hooks/` harbors our git pre-commit hooks that enforce coding style.

## Endpoints

### `GET /programs/:id`, `GET /userData/:id`

Get an `User` or `Program` document with UID `:id` in JSON form.

Example nominal response:

```json
// example /programs/ response
// response code 200
{
    "code": "def howdy():\n  print('hi')\n",
    "dateCreated": "2019-12-14T19:14:08.457733Z",
    "language": "python",
    "name": "Program name",
    "thumbnail": 0
}

// example /userData/ response
// response code 200
{
    "displayName": "Joe Bruin",
    "photoName": "",
    "mostRecentProgram": "PROGHASH",
    "programs": [
        "HASH0",
        "HASH1",
        "HASH2"
    ],
    "classes": null
}
```

### `PUT /programs/:id`, `PUT /userData/:id`

Updates the user or program document with uid `:id` with the data provided in the request body. The `/programs/` endpoint takes an array of `Program`s in the request body, while the `/userData/` endpoint takes one or more `User` fields. These endpoints, actually, **aren't properly implemented yet**. Here's what the requests should look like:

Example Request:

```json
// sample user request body
{
    "displayName": "TLA Dev Team"
}

// sample program request body
{
    "HASH0": {
        "name": "my updated program name",
        "code": "print('here\'s my new code for the program!')"
    },
    "HASH1": {
        "name": "another program, this time just updating the name."
    }
}
```

Example nominal response: `200`

### `POST /programs/`

Creates a new program document associated to a user with information as supplied through the request body:

```json
{
	"uid": "my cool user ID!",
	"name": "my neato processing program!",
	"language": "processing",
	"thumbnail": 25
}
```

Example nominal response:

```json
// response code: 200
{
    "displayName": "J Bruin",
    "photoName": "",
    "mostRecentProgram": "",
    "programs": [
        "HASH0",
        "HASH1",
        "HASH2"
    ],
    "classes": null
}
```

### `POST /userData/`

Creates a new `User` document with the default programs. There are no special requirements for the request body.

Example nominal response:

```json
// response code: 200
{
    "displayName": "J Bruin",
    "photoName": "",
    "mostRecentProgram": "",
    "programs": [
        "HASH0",
        "HASH1",
        "HASH2"
    ],
    "classes": null
}
```

### `DELETE /programs/:id`

Delete the program with uid `:id` from the user with uid `:uid`, as provided in the request body.

```json
{
    "uid": "my cool program ID"
}
```

Example nominal response: `200`