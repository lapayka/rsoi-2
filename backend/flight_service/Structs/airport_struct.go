package FS_structs

type Airport struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	City    string `json:"city"`
	Country string `json:"country"`
}
