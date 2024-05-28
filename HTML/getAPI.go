package never

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func getRelation(artistID int, w http.ResponseWriter, r *http.Request) Relation {
	// Fetch the artist's relations information using the artist ID
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/relation/" + strconv.Itoa(artistID))
	if err != nil {
		HandleInternalError(w, r)
		return Relation{}
	}
	defer resp.Body.Close()

	var relation Relation
	err = json.NewDecoder(resp.Body).Decode(&relation)
	if err != nil {
		HandleInternalError(w, r)
		return Relation{}
	}
	return relation
}

func getLocation(artistID int, w http.ResponseWriter, r *http.Request) Location {
	// Fetch the artist's location information using the artist ID
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/locations/" + strconv.Itoa(artistID))
	if err != nil {
		HandleInternalError(w, r)
		return Location{}
	}
	defer resp.Body.Close()

	var location Location
	err = json.NewDecoder(resp.Body).Decode(&location)
	if err != nil {
		http.Redirect(w, r, "Error 500", http.StatusInternalServerError)
		return Location{}
	}
	return location
}

func getDates(artistID int, w http.ResponseWriter, r *http.Request) Date {
	// Fetch the artist's relations information using the artist ID
	resp, err := http.Get("https://groupietrackers.herokuapp.com/api/dates/" + strconv.Itoa(artistID))
	if err != nil {
		HandleInternalError(w, r)
		return Date{}
	}
	defer resp.Body.Close()

	var date Date
	err = json.NewDecoder(resp.Body).Decode(&date)
	if err != nil {
		HandleInternalError(w, r)
		return Date{}
	}
	return date
}
