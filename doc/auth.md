# Auth Design

## AuthN Types

1. Password
2. OAuth2.0
3. Passwordless (Email OTP)

## General Flow

The end goal of auth is always the same, create a session and store in database. 
Everytime the client tries to access a protected API, it requires the session cookie.
Each request can use the session cookie to extract the user, and perform actions based on
that user.

```
Password login     OAuth login     Passwordless login
      |                 |                 |
      v                 v                 v
verify password   verify with Google   verify magic link / OTP
      |                 |                 |
      v                 v                 v
  get userID        get userID        get userID
      |                 |                 |
      +--------+--------+-----------------+
               |
               v
         createSession
```

## Password-based

1. Initial set up: Ask the user for user & password as minimum. Store password using bcrypt, and store in db.
2. On login, the client asks for the password and sends the same hash to server.
3. In db, fetch the user id using the username.
4. Using the user id, fetch the stored password hash. (Step3&4 is easily done with SQL JOIN)
5. Use the `bcrypt.compare('password', stored_hash)` to see if there's a match.
6. If matched, the server returns the session cookie back to client

## Google 

1. Generate a random `state` string. Redirect the user to Google's log in page
2. Google does the validation, and asks user to approve the application. Google calls the `auth/callback` endpoint
with the code and state.
3. Server validates the returned state matches (step 1 and step 2). Using the given token, the server asks Google whose
code it belongs to.
4. Google returns the id_token. In DB, id_token can be looked up to find the matching user id.
5. The server creates a new user if doesn't exist
6. Server generates a new session cookie and returns back to client.

### Flow

```
Browser                        Your Go Server                    Google
  |                                  |                               |
  |-- GET /auth/login -------------> |                               |
  |                                  | generate random `state`       |
  |                                  | store state in temp cookie    |
  | <-- redirect to Google ----------|                               |
  |                                                                  |
  |-- user logs in & approves ---------------------------------->    |
  |                                                                  |
  | <-- redirect to /auth/callback?code=XXX&state=YYY ----------     |
  |                                  |                               |
  |-- GET /auth/callback ----------> |                               |
  |                                  | verify state matches cookie   |
  |                                  |-- POST /token (code) -------> |
  |                                  | <-- access_token, id_token -- |
  |                                  | decode id_token → sub, email  |
  |                                  | upsert user in DB             |
  |                                  | create session token in DB    |
  | <-- Set-Cookie: session=TOKEN ---|                               |
  |                                  |                               |
  |-- GET /api/me (cookie) --------> |                               |
  |                                  | look up session token in DB   |
  |                                  | get user_id → load user       |
  | <-- { email, name } -------------|                               |
```

## AuthZ

### RBAC

Soon (TM)

## Database schema

Please see [database_schema.md](database_schema.md#AuthN)

### More learning opportunities
1. How to prevent server from getting bombarded with new user requests
2. For password-based, how to implement max retries.