package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gorilla/feeds"
	"html/template"
	"net/http"
	"net/url"
	"time"
)

type houseListing struct {
	Id           string   `json:"_id"`
	Address1     string   `json:"Address1"`
	Address2     string   `json:"Address2"`
	ContractDate string   `json:"Contract Date"`
	ListPrice    string   `json:"List"`
	MLSNumber    string   `json:"MLS#"`
	Tags         []string `json:"Undefined"`
	Pictures     []picture
	MLSListing   MLSListing `json:"-"`
}

type picture struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

var descriptionTemplate = template.Must(template.New("desc").Parse(
	`
	<p>
	{{.remarks}}
	</p>
	<iframe width="600" height="450" frameborder="0" style="border:0" src="https://www.google.com/maps/embed/v1/place?key={{.maps_api_key}}&q={{.latitude}},{{.longitude}}" allowfullscreen></iframe>
	<p/>
	<a href="{{.mls_link}}">MLS</a>
	<p/>
	{{range .photos}}
		<img src="{{.}}"/>
	{{end}}
	`))

func getHouses(w http.ResponseWriter, _ *http.Request) {
	resp, err := http.Get("https://www.mongohouse.com/api/newlistings?query=true&price_min=$300,000&price_max=$750,000&list_day_back=120&bedrooms=2&washrooms=1&ownershiptype=freehold&openhouse=undefined&south=43.653101797558975&west=-79.4750231366699&north=43.71008371751932&east=-79.28756891303709")
	if err != nil || resp == nil || resp.Body == nil {
		return
	}
	defer resp.Body.Close()
	var listings []houseListing
	json.NewDecoder(resp.Body).Decode(&listings)
	feed := &feeds.Feed{
		Title: "House listings",
		Link:  &feeds.Link{Href: "http://mongohouse.com"},
	}
	for _, listing := range listings {
		mls, err := getMLS(listing.MLSNumber)
		if err != nil {
			mls = &MLSListing{PublicRemarks: err.Error()}
		}
		mlsUrl := fmt.Sprintf("https://www.realtor.ca%s", mls.RelativeDetailsURL)
		var buf bytes.Buffer
		var photos []string
		for _, photo := range mls.Property.Photos {
			photos = append(photos, photo.MedResPath)
		}
		descriptionTemplate.Execute(&buf, map[string]interface{}{
			"maps_api_key": mapsApiKey,
			"longitude":    mls.Property.Address.Longitude,
			"latitude":     mls.Property.Address.Latitude,
			"remarks":      mls.PublicRemarks,
			"mls_link":     mlsUrl,
			"photos":       photos,
		})
		listingDate, _ := time.Parse("01/02/2006", listing.ContractDate)
		item := &feeds.Item{
			Title:       fmt.Sprintf("%s - %s", listing.Address1, listing.ListPrice),
			Link:        &feeds.Link{Href: fmt.Sprintf("https://mongohouse.com/newlistings/%s", listing.Id)},
			Created:     listingDate,
			Description: buf.String(),
		}
		feed.Items = append(feed.Items, item)
	}
	rss, _ := feed.ToRss()
	w.Write([]byte(rss))
}

type MLSListing struct {
	Building struct {
		BathroomTotal string `json:"BathroomTotal"`
		Bedrooms      string `json:"Bedrooms"`
		StoriesTotal  string `json:"StoriesTotal"`
		Type          string `json:"Type"`
	} `json:"Building"`
	Property struct {
		Address struct {
			AddressText string `json:"AddressText"`
			Longitude   string `json:"Longitude"`
			Latitude    string `json:"Latitude"`
		} `json:"Address"`
		Photos []struct {
			HighResPath string `json:"HighResPath"`
			MedResPath  string `json:"MedResPath"`
			LowResPath  string `json:"LowResPath"`
		} `json:"Photo"`
	} `json:"Property"`
	PublicRemarks      string `json:"PublicRemarks"`
	RelativeDetailsURL string `json:"RelativeDetailsURL"`
}

type mlsResponse struct {
	Results []MLSListing `json:"Results"`
}

func newId() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func getMLS(id string) (*MLSListing, error) {
	postData := url.Values{}
	postData.Set("CultureId", "1")
	postData.Set("ApplicationId", "1")
	postData.Set("ReferenceNumber", id)
	postData.Set("IncludeTombstones", "1")
	postData.Set("Version", "6.0")
	postData.Set("GUID", newId())
	fmt.Println("FETCHING:", id)
	resp, err := http.PostForm("https://api2.realtor.ca/Listing.svc/PropertySearch_Post", postData)
	if err != nil {
		fmt.Println("ERR:", err)
		return nil, err
	}
	fmt.Println("Got ", id)
	var listings mlsResponse
	json.NewDecoder(resp.Body).Decode(&listings)
	if len(listings.Results) > 0 {
		return &listings.Results[0], nil
	}
	return nil, fmt.Errorf("No result found")
}
