package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/feeds"
	"net/http"
	"time"
)

type houseListing struct {
	Id           string `json:"_id"`
	Address1     string `json:"Address1"`
	Address2     string `json:"Address2"`
	ContractDate string `json:"Contract Date"`
	ListPrice    string `json:"List"`
	MLSNumber    string `json:"MLS#"`
	Pictures     []picture
}

type picture struct {
	Url         string `json:"url"`
	Description string `json:"description"`
}

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
		listingDate, _ := time.Parse("01/02/2006", listing.ContractDate)
		item := &feeds.Item{
			Title:   fmt.Sprintf("%s - %s", listing.Address1, listing.ListPrice),
			Link:    &feeds.Link{Href: fmt.Sprintf("https://mongohouse.com/newlistings/%s", listing.Id)},
			Created: listingDate,
		}
		feed.Items = append(feed.Items, item)
	}
	rss, _ := feed.ToRss()
	w.Write([]byte(rss))
}
