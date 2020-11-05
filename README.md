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
