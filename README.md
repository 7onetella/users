# Purpose
There is a need for good secure authentication. I have decided to implement an authentication server to be able to
understand JWT, TOTP and FIDO2 WebAuthn. 

# Live Demo Site
Go to [My Accounts](https://accounts.7onetella.net/accounts/)
- register your own accounts please

# Demo
![](/assets/auth.gif)

# API documentation
[Redoc](https://accounts.7onetella.net/accounts/redoc.html)

# Work In Progress
There is a lot to be done in order to make the authentication even more secure. I will be doing some internal clean up
of code in near future.

# Future enhancement
- Add source IP check against previously recorded source IPs
- Add backoff period when password auth or totp auth fails three times in a row
- Add browser agent check against previously recorded browser agents
- Add OAuth2 support
- Add Swagger documentation

# Acknowledgement
- [Duo WebAuthn](https://github.com/duo-labs/webauthn)
- [Bootstrap 4](https://github.com/twbs/bootstrap)
- [EmberJS](https://github.com/emberjs/ember.js)
- [QR Code](https://github.com/rsc/qr)
- [TOTP](https://github.com/xlzd/gotp)
- [JWT](http://github.com/dgrijalva/jwt-go)
- [Gin](http://github.com/gin-gonic/gin)
- [SQLX](http://github.com/jmoiron/sqlx)

 ### Introduction
 The design of User API is around web apps. However, there is nothing that prevents anyone from leveraging
 oauth2 with User API in native apps.

 ### Context
 Security is the most important thing in apps. Password auth is insecure as compared to TOTP or WebAuthn.
 We want to give users option of enabling additional authentication mechanism.

 ### Forces that impact the design
 * Avoid using cookies to store auth related tokens that represent successful authentication
 * Give consideration for javascript client authentication library
 * Javascript client library support lacks WebAuthn

 ### Design Decision
 Cookies are not necessary bad but cookies are sent every time browser contacts the server.
 The decision to send auth token should be delegated to javascript auth client library.
 User API does not send auth related cookies. JWT Token is sent in response JSON payload instead.

