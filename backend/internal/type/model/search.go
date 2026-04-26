package model

type Search struct {
	Abstract       string `json:"Abstract"`
	AbstractText   string `json:"AbstractText"`
	AbstractSource string `json:"AbstractSource"`
	AbstractURL    string `json:"AbstractURL"`
	RelatedTopics  []struct {
		Text     string `json:"Text"`
		FirstURL string `json:"FirstURL"`
	} `json:"RelatedTopics"`
}
