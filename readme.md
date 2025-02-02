# p-chat

**_`p-chat`_** is a chat application that allows users to create chat rooms and chat with other users in real-time.


## Client Connection

The client connects to the server using a WebSocket connection. Chat endpoints are guarded by a oAuth JWT token, which is obtainable by user/password credentials for limited time, scope and renew ability restrictions.
![Alt text](https://cdn.discordapp.com/attachments/341254180582981632/1335613734483132528/image.png?ex=67a0ceb8&is=679f7d38&hm=9f3754417c1eef591ab479dfd269f121ae02de27e2811f2f43515eab5b2b00f8&)


## Environment Variables

Represents variables that are used to modify software behavior and establish connection with proper modules like database.
```
// .env.example
JWT_SECRET=your-secret-key-here
ACCESS_TOKEN_SESSION_MINUTES=15
REFRESH_TOKEN_SESSION_HOURS=1

PSQL_HOST=
PSQL_USER=
PSQL_PASSWORD=
PSQL_DB=
```

## API Documentation

### 1. POST /session/authorize

Logs in a user and returns a session token.

#### Body:

| name     | description            |
|----------|------------------------|
| email    | string email username  |
| password | sha256 password string |


#### Returns: 

```json
status 200:
{
    status: 'Success', // or 'Error'
    data: {
        access_token: 'complex token',
        refresh_token: 'complex token'
        access_token_valid_to: 'date time (ISO)',
        refresh_token_valid_to: 'date time (ISO)',
    },
}
status 401:
{}
```

### 2. POST /session/register

Refreshes a session token.

#### Body:

| name     | description            |
|----------|------------------------|
| email    | string email username  |
| password | sha256 password string |

#### Returns: 

```json
status 200:
{
    status: 'Success', // or 'Error'
    data: {
        VerificationToken: 'complex token',
        message: 'User registered successfully',
    },
}
status 401:
{}
```

### 3. GET /session/emailVerify?verifyToken='complex token'

Verifies a user's email address.

#### Returns: 

```json
status 200:
{
  status: 'Success', // or 'Error'
  data: {
    access_token: 'complex token',
    refresh_token: 'complex token'
    access_token_valid_to: 'date time (ISO)',
    refresh_token_valid_to: 'date time (ISO)',
  },
}
status 401:
{}
```

## Database Schema


![Alt text](https://cdn.discordapp.com/attachments/341254180582981632/1335677491448254510/image.png?ex=67a10a19&is=679fb899&hm=d7ccadf0846ceccdd097a43e83a6833cb55928beaec825b2cce22d39504d6c7e&)


## Testing

```
$ go test ./...
```

## Build

```
$ go build main.go
```

## Run

```
$ ./main
```

## License

MIT License

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.