package domain

type Source struct {
	Id       int    `json:"id"`
	Category string `json:"category"`
	Status   string `json:"status"`
}
