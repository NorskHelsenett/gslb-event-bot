package routes

import "net/http"

const (
	ROOT = "/"

	SUBSCRIBE      = ROOT + "subscribe"
	POST_SUBSCRIBE = http.MethodPost + " " + SUBSCRIBE
)
