package never

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

type HomePageData struct {
	Artists     []Artist
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

// HandleRequest handles incoming HTTP requests and generates a response
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Check if the request path is not the root or index.html
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		HandleNotFound(w, r)
		return
	}
	// Check if the request method is not GET
	if r.Method != http.MethodGet {
		HandleMethod(w, r)
		return
	}

	// Extract the query parameter from the URL
	query := r.URL.Query().Get("query")

	// Make a GET request to fetch artists data from an external API
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists")
	if err != nil {
		HandleInternalError(w, r)
		return
	}
	defer resp.Body.Close()

	var artists []Artist
	// Decode the JSON response into the artists slice
	err = json.NewDecoder(resp.Body).Decode(&artists)
	if err != nil {
		HandleInternalError(w, r)
		return
	}

	var filteredArtists []Artist
	var suggestions []string
	if query != "" {
		query = strings.ToLower(query)
		// Iterate through each artist to filter based on the query
		for _, artist := range artists {
			// Fetch additional details for the artist
			artistWithInfo, err := fetchArtistDetails(artist)
			if err != nil {
				continue
			}
			// Check if the artist matches the query
			matches := containsQuery(artistWithInfo, query)
			if len(matches) > 0 {
				filteredArtists = append(filteredArtists, artist)
				suggestions = append(suggestions, matches...)
			}
		}
		// Limit the number of suggestions to 10
		if len(suggestions) > 10 {
			suggestions = suggestions[:10]
		}
	} else {
		// If no query, include all artists and suggest them
		filteredArtists = artists
		for _, artist := range artists {
			suggestions = append(suggestions, artist.Name+" - artist/band")
		}
	}

	// Prepare data to be passed to the template
	data := struct {
		Artists     []Artist
		Suggestions []string
	}{
		Artists:     filteredArtists,
		Suggestions: suggestions,
	}

	// Parse the HTML template file
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		HandleInternalError(w, r)
		return
	}

	// Execute the template with the provided data and write to the response writer
	err = tmpl.Execute(w, data)
	if err != nil {
		HandleInternalError(w, r)
		return
	}
}

// fetchArtistDetails fetches additional details for an artist
func fetchArtistDetails(artist Artist) (ArtistWithInfo, error) {
	var artistWithInfo ArtistWithInfo
	artistWithInfo.Artist = artist

	// Fetch locations for the artist
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/locations/" + strconv.Itoa(artist.ID))
	if err != nil {
		return artistWithInfo, err
	}
	defer resp.Body.Close()
	var location Location
	err = json.NewDecoder(resp.Body).Decode(&location)
	if err != nil {
		return artistWithInfo, err
	}
	artistWithInfo.Locations = location.LocationS

	// Fetch concert dates for the artist
	resp, err = http.Get("https://groupietrackers.herokuapp.com/api/dates/" + strconv.Itoa(artist.ID))
	if err != nil {
		return artistWithInfo, err
	}
	defer resp.Body.Close()
	var date Date
	err = json.NewDecoder(resp.Body).Decode(&date)
	if err != nil {
		return artistWithInfo, err
	}
	artistWithInfo.Dates = date.ConcertDates

	return artistWithInfo, nil
}

// containsQuery checks if an artist matches the search query and returns relevant matches
func containsQuery(artist ArtistWithInfo, query string) []string {
	var matches []string

	// Check if the artist's name contains the query
	if strings.Contains(strings.ToLower(artist.Name), query) {
		// If it matches, add it to the matches slice
		matches = append(matches, artist.Name+" - artist/band")
	}

	// Check if any member of the artist contains the query
	for _, member := range artist.Members {
		if strings.Contains(strings.ToLower(member), query) {
			// If it matches, add it to the matches slice
			matches = append(matches, artist.Name+" - member: "+member)
		}
	}

	// Check if the artist's creation date matches the query
	if strings.Contains(strings.ToLower(strconv.Itoa(artist.CreationDate)), query) {
		// If it matches, add it to the matches slice
		matches = append(matches, artist.Name+" - creation date: "+strconv.Itoa(artist.CreationDate))
	}

	// Check if the artist's first album matches the query
	if strings.Contains(strings.ToLower(artist.FirstAlbum), query) {
		// If it matches, add it to the matches slice
		matches = append(matches, artist.Name+" - first album: "+artist.FirstAlbum)
	}

	// Check if any location associated with the artist matches the query
	for _, location := range artist.Locations {
		if strings.Contains(strings.ToLower(location), query) {
			// If it matches, add it to the matches slice
			matches = append(matches, artist.Name+" - location: "+location)
		}
	}

	// Check if any concert date associated with the artist matches the query
	for _, date := range artist.Dates {
		if strings.Contains(strings.ToLower(date), query) {
			// If it matches, add it to the matches slice
			matches = append(matches, artist.Name+" - concert date: "+date)
		}
	}

	// Return the list of matches
	return matches
}
// HandleRequest2 handles a specific type of request for detailed artist information
func HandleRequest2(w http.ResponseWriter, r *http.Request) {
	// Extract the artist ID from the query parameters
	artistID := r.URL.Query().Get("id")

	// Convert the artistID to an integer
	id, err := strconv.Atoi(artistID)
	if err != nil {
		HandleNotFound(w, r)
		return
	}

	// Check if the ID is within a valid range
	if id > 52 || id <= 0 {
		HandleNotFound(w, r)
		return
	}

	// Fetch detailed information about the artist using the provided ID
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/artists/" + artistID)
	if err != nil {
		HandleInternalError(w, r)
		return
	}
	defer resp.Body.Close()

	// Decode the artist's detailed information from the response
	var artist Artist
	err = json.NewDecoder(resp.Body).Decode(&artist)
	if err != nil {
		http.Error(w, "Failed to decode artist data", http.StatusBadGateway)
		return
	}

	// Get additional information such as location, relation, and concert dates
	location := getLocation(artist.ID, w, r)
	relation := getRelation(artist.ID, w, r)
	concertDate := getDates(artist.ID, w, r)

	// Create a new instance of ArtistWithInfo containing all gathered information
	artistWithInfo := ArtistWithInfo{
		Artist:    artist,
		Locations: location.LocationS,
		Dates:     concertDate.ConcertDates,
		Relations: relation.DatesLocations,
	}

	// Load the info.html template
	tmpl, err := template.ParseFiles("templates/info.html")
	if err != nil {
		HandleInternalError(w, r)
		return
	}

	// Execute the template with the artist's detailed information
	err = tmpl.Execute(w, artistWithInfo)
	if err != nil {
		HandleInternalError(w, r)
		return
	}
}
