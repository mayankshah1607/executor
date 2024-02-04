package server

type StatsResponse struct {
	Waiting []string `json:"waiting"`
	Running []string `json:"running"`
}
