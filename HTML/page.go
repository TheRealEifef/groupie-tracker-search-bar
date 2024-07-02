package never

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type HomePageData struct {
	Artists     []ArtistWithInfo
	Suggestions []string
}

type Artist struct {
	ID           int      `json:"id"`
	Image        string   `json:"image"`
	Name         string   `json:"name"`
	Members      []string `json:"members"`
	CreationDate int      `json:"creationDate"`
	FirstAlbum   string   `json:"firstAlbum"`
	Locations    string   `json:"locations"`
	ConcertDates string   `json:"dates"`
	Relations    string   `json:"datesLocations"`
}
type Location struct {
	ID        int      `json:"id"`
	LocationS []string `json:"locations"`
}
type Date struct {
	ID           int      `json:"id"`
	ConcertDates []string `json:"dates"`
}
type Relation struct {
	ID             int                 `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}
type ArtistWithInfo struct {
	Artist
	Locations []string            `json:"locations"`
	Dates     []string            `json:"dates"`
	Relations map[string][]string `json:"datesLocations"`
}

func HandleRequest(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		HandleNotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		HandleMethod(w, r)
		return
	}

	query := r.FormValue("query")

	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		HandleInternalError(w, r)
		return
	}
	defer resp.Body.Close()

	var artists []Artist
	err = json.NewDecoder(resp.Body).Decode(&artists)
	if err != nil {
		HandleInternalError(w, r)
		return
	}

	var artistWithInfo []ArtistWithInfo
	var filteredArtists []ArtistWithInfo
	var suggestions []string
	exactMatches := []string{}
	partialMatches := []string{}
	if query != "" {
		query = strings.ToLower(query)
		if len(artistWithInfo) > 0 {
			for _, artist := range artistWithInfo {
				if containsQuery(artist, query) {
					filteredArtists = append(filteredArtists, artist)
					if strings.EqualFold(artist.Name, query) {
						exactMatches = append(exactMatches, artist.Name)
					} else {
						partialMatches = append(partialMatches, artist.Name)
					}
				}
			}
		}
		suggestions = append(exactMatches, partialMatches...)
		if len(suggestions) > 10 {
			suggestions = suggestions[:10]
		}
	} else {
		filteredArtists = artistWithInfo
		for _, artist := range artists {
			suggestions = append(suggestions, artist.Name)
		}
	}

	data := HomePageData{
		Artists:     filteredArtists,
		Suggestions: suggestions,
	}

	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		HandleInternalError(w, r)
		return
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		HandleInternalError(w, r)
		return
	}
}

func containsQuery(artist ArtistWithInfo, query string) bool {
	// Convert the query to lowercase
	queryLower := strings.ToLower(query)

	// Check if the query exactly matches the artist's name
	if strings.Contains(strings.ToLower(artist.Name), queryLower) {
		return true
	}

	// Check if half the letters in the query match the artist's name
	if matchesHalf(artist.Name, queryLower) {
		return true
	}

	// Check if the query exactly matches any of the artist's members
	for _, member := range artist.Members {
		if strings.Contains(strings.ToLower(member), queryLower) {
			return true
		}
	}

	// Check if half the letters in the query match any of the artist's members
	for _, member := range artist.Members {
		if matchesHalf(member, queryLower) {
			return true
		}
	}

	// Convert the artist's creation date to a string in the desired format
	creationDateStr := time.Date(artist.CreationDate/10000, time.Month(artist.CreationDate/100%100), artist.CreationDate%100, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	// Check if the query exactly matches the artist's creation date
	if strings.Contains(strings.ToLower(creationDateStr), queryLower) {
		return true
	}

	// Check if half the letters in the query match the artist's creation date
	if matchesHalf(creationDateStr, queryLower) {
		return true
	}

	// Check if the query exactly matches the artist's first album
	if strings.Contains(strings.ToLower(artist.FirstAlbum), queryLower) {
		return true
	}

	// Check if half the letters in the query match the artist's first album
	if matchesHalf(artist.FirstAlbum, queryLower) {
		return true
	}

	locationsStr := strings.Join(artist.Locations, " ")
	if strings.Contains(strings.ToLower(locationsStr), queryLower) {
		return true
	}
	if matchesHalf(locationsStr, queryLower) {
		return true
	}

	// Check if the query exactly matches the artist's concert dates
	if strings.Contains(strings.ToLower(artist.ConcertDates), queryLower) {
		return true
	}

	// Check if half the letters in the query match the artist's concert dates
	if matchesHalf(artist.ConcertDates, queryLower) {
		return true
	}

	return false
}

func matchesHalf(str, query string) bool {
	// Count the number of matching characters
	matchCount := 0
	for i := range query {
		if i < len(str) && str[i] == query[i] {
			matchCount++
		}
	}

	// Check if half the letters in the query match the string
	return float64(matchCount) >= float64(len(query))/2
}

func getArtistWithInfo(artistID int, w http.ResponseWriter, r *http.Request) ArtistWithInfo {
	// Fetch the artist's information from the API
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists/" + strconv.Itoa(artistID))
	if err != nil {
		HandleInternalError(w, r)
		return ArtistWithInfo{}
	}
	defer resp.Body.Close()

	var artist Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	if err != nil {
		HandleInternalError(w, r)
		return ArtistWithInfo{}
	}

	// Fetch the artist's locations, dates, and relations
	location := getLocation(artistID, w, r)
	date := getDates(artistID, w, r)
	relation := getRelation(artistID, w, r)

	return ArtistWithInfo{
		Artist:    artist,
		Locations: location.LocationS,
		Dates:     date.ConcertDates,
		Relations: relation.DatesLocations,
	}
}

func HandleRequest2(w http.ResponseWriter, r *http.Request) {
	// Extract the artist ID from the query parameters
	artistID := r.URL.Query().Get("id")

	// Convert the artistID to an integer
	id, err := strconv.Atoi(artistID)
	if err != nil {
		HandleInternalError(w, r)
		return
	}

	// Fetch the artist's information
	artistInfo := getArtistWithInfo(id, w, r)

	// Render the artist's information in the template
	tmpl, err := template.ParseFiles("templates/info.html")
	if err != nil {
		HandleInternalError(w, r)
		return
	}

	err = tmpl.Execute(w, artistInfo)
	if err != nil {
		HandleInternalError(w, r)
		return
	}
}
