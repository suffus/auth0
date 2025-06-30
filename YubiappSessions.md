Make Yubiapp support sessions by adding endpoints supporting the issuing of JWT-based access tokens and refresh tokens.  

The supported workflow is as follows:

1. The user enters his yubikey into a login dialog on the frontend app.
2. The frontend app sends the code (and maybe the username, appname, and a permission) to the yubiapp at a new endpoint /auth/session via a POST query. 
3. Yubiapp returns JSON as in /auth/device, plus a session identifier, an access token (which expires after no more than a configurable number of minutes, usually no more than 15 minutes). and a refresh token (which will not expire during the anticipated use of the web app by the user).  The access token is a JWT which includes a session id.  The refresh token is a JWT which includes the session id, plus a counter which indicates the number of access tokens that have been issued during the session.
4. The frontend web app may use the access token to authenticate most methods on the YubiAuth backend that only involve reading data, but not those that involve creating or modifying data.  Whether they can be used to perform an action will be configurable on a per-action basis.
5. If the web app needs to issue a new access token (for example if its existing one is due to expire), it may send the refresh token to a new endpoint /auth/session/refresh/:session_id via a POST query, which will increment a counter in the stored session, and forge new access and refresh token for it.  All existing access or refresh tokens will then be invalidated for this session.

Sessions will be stored in a an in-memory Redis database for speed of access, and the entries will store the session identifier (a uuid), an access counter which will be incremented each time an access token is used to authenticate an api call, and a refresh count, which counts the number of access tokens that have been issued for this session (via /auth/session/refresh).  It will also store the device id of the device that was authenticated, the user id of the device's user at that time, the time of creation of the session, a flag to say whether the session is still valid, and the date of expiry of the latest refresh token.

The idea here is to build a system which does not require the user to authenticate every call with a device, but allows limited read access by authenticating once, and then using an access token to make further calls.  We do want to be reasonably sure that the session was originally authorized by the user, and that sessions can't be shared.  To this end, the access tokens will be relatively short-lived, and refresh tokens can only be used once because their access token counter must always equal the number of access tokens that have been issued.

Implementation details:

create the endpoints /auth/session and /auth/session/refresh, and code to connect to the redis database to store sessions (redis config should be in the config file of course).

implement the creation of refresh tokens and access tokens as JWT tokens described above.

