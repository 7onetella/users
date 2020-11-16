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

## Table of Content

* [What is Lorem Ipsum](#-what-is-lorem-ipsum)
* [Where does  it come from](#wiki-Where does  it come from)
* [Why do we use it](#wiki-Why do we use it)
* [Where can I get some](#wiki-Where can I get some)

## <a name="What is Lorem Ipsum"/> What is Lorem Ipsum

Lorem Ipsum is simply dummy text of the printing and typesetting industry. Lorem Ipsum has been the industry's standard dummy text ever since the 1500s, when an unknown printer took a galley of type and scrambled it to make a type specimen book. It has survived not only five centuries, but also the leap into electronic typesetting, remaining essentially unchanged. It was popularised in the 1960s with the release of Letraset sheets containing Lorem Ipsum passages, and more recently with desktop publishing software like Aldus PageMaker including versions of Lorem Ipsum.

## <a name="Where does  it come from"/> Where does  it come from

Contrary to popular belief, Lorem Ipsum is not simply random text. It has roots in a piece of classical Latin literature from 45 BC, making it over 2000 years old. Richard McClintock, a Latin professor at Hampden-Sydney College in Virginia, looked up one of the more obscure Latin words, consectetur, from a Lorem Ipsum passage, and going through the cites of the word in classical literature, discovered the undoubtable source. Lorem Ipsum comes from sections 1.10.32 and 1.10.33 of "de Finibus Bonorum et Malorum" (The Extremes of Good and Evil) by Cicero, written in 45 BC. This book is a treatise on the theory of ethics, very popular during the Renaissance. The first line of Lorem Ipsum, "Lorem ipsum dolor sit amet..", comes from a line in section 1.10.32.
The standard chunk of Lorem Ipsum used since the 1500s is reproduced below for those interested. Sections 1.10.32 and 1.10.33 from "de Finibus Bonorum et Malorum" by Cicero are also reproduced in their exact original form, accompanied by English versions from the 1914 translation by H. Rackham.

## <a name="Why do we use it"/> Why do we use it

It is a long established fact that a reader will be distracted by the readable content of a page when looking at its layout. The point of using Lorem Ipsum is that it has a more-or-less normal distribution of letters, as opposed to using 'Content here, content here', making it look like readable English. Many desktop publishing packages and web page editors now use Lorem Ipsum as their default model text, and a search for 'lorem ipsum' will uncover many web sites still in their infancy. Various versions have evolved over the years, sometimes by accident, sometimes on purpose (injected humour and the like).

## <a name="Where can I get some"/> Where can I get some

There are many variations of passages of Lorem Ipsum available, but the majority have suffered alteration in some form, by injected humour, or randomised words which don't look even slightly believable. If you are going to use a passage of Lorem Ipsum, you need to be sure there isn't anything embarrassing hidden in the middle of text. All the Lorem Ipsum generators on the Internet tend to repeat predefined chunks as necessary, making this the first true generator on the Internet. It uses a dictionary of over 200 Latin words, combined with a handful of model sentence structures, to generate Lorem Ipsum which looks reasonable. The generated Lorem Ipsum is therefore always free from repetition, injected humour, or non-characteristic words etc.
