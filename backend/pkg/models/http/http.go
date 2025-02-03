package http

type Body struct {
	Key string `json:"key" binding:"required"`
	Val string `json:"value" binding:"required"`
}
