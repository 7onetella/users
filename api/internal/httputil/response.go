package httputil

func (rh RequestHandler) WriteCORSHeader() {
	w := rh.Context.Writer
	w.Header().Add("Access-Control-Allow-Origin", "*")
}

func (rh RequestHandler) SetContentTypeJSON() {
	c := rh.Context
	c.Header("Content-Type", "application/json")
}

//func (rh RequestHandler) GrantAccess(claimKey, userID, jwtSecret string, ttl time.Duration) {
//	c := rh.Context
//	tokenString, expTime, err := EncodeToken(claimKey, userID, jwtSecret, ttl)
//	if err != nil {
//		log.Println("encoding error")
//		c.AbortWithError(http.StatusInternalServerError, err)
//		return
//	}
//
//	log.Println("Signin successful")
//
//	token := handlers.AuthToken{
//		Token:      tokenString,
//		Expiration: expTime.Unix(),
//	}
//	c.JSON(200, token)
//}
