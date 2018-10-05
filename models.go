package main

// Show is the tvmaze Show structure
type Show struct {
	ID           int
	URL          string
	Name         string
	Language     string
	Genre        []string
	Status       string
	Runtime      int
	Premiered    string
	OfficialSite string
	Summary      string
}

// Episode is the structure of the API episode
type Episode struct {
	ID     int
	URL    string
	Name   string
	Season int
	Number int
}

// JResponse is the first JSON struct we get from the API
type JResponse struct {
	Score float64
	Show  Show
}
