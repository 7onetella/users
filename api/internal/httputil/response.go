package httputil

func (rh RequestHanlder) WriteCORSHeader() {
	w := rh.Context.Writer
	w.Header().Add("Access-Control-Allow-Origin", "*")
}

func (rh RequestHanlder) SetContentTypeJSON() {
	c := rh.Context
	c.Header("Content-Type", "application/json")
}
