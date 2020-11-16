# Purpose
There is a need for good secure authentication. This project is a very light implementation of authentication server.
Three different types of authentication method have been implemented. 

|      1st Factor             |          2nd Factor       |     Security                      |
|-----------------------------|---------------------------|-----------------------------------|
|      Password               |             None          | Weak - password can be guessed    |
|      Password               |             TOTP          | Medium - prone to phishing attack |
|      Password               |             WebAuthn      | Strong                            |

[U2F](https://en.wikipedia.org/wiki/Universal_2nd_Factor) key with WebAuthn protocol improves web security. It is possible to
go password-less. However, it will be awhile before mass adoption takes place.
  
# Screen Capture
![](/assets/auth.gif)

# Live Demo Site
Go to [Demo Site](https://accounts.7onetella.net/accounts/)

\* **register your own accounts please**

# API documentation
[Go Here](https://accounts.7onetella.net/accounts/redoc.html)

# Future enhancement
- [x] ~~Add OAuth2 support~~
- [x] ~~Add Swagger documentation~~
- [ ] Add source IP check against previously recorded source IPs
- [ ] Add backoff period when password auth or totp auth fails three times in a row
- [ ] Add browser agent check against previously recorded browser agents

# Acknowledgement
- [Duo WebAuthn](https://github.com/duo-labs/webauthn)
- [Bootstrap 4](https://github.com/twbs/bootstrap)
- [EmberJS](https://github.com/emberjs/ember.js)
- [QR Code](https://github.com/rsc/qr)
- [TOTP](https://github.com/xlzd/gotp)
- [JWT](http://github.com/dgrijalva/jwt-go)
- [Gin](http://github.com/gin-gonic/gin)
- [SQLX](http://github.com/jmoiron/sqlx)
