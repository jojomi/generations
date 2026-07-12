package main

import (
	"encoding/json"
	"html/template"
	"strings"

	"github.com/jojomi/generations"
)

type SearchField struct {
	Label string `json:"l"`
	Value string `json:"v"`
}

type SearchEntry struct {
	Name   string        `json:"n"`
	URL    string        `json:"u"`
	Fields []SearchField `json:"f"`
}

// personDetailsURL returns the relative link to a person's detail page.
func personDetailsURL(person generations.Person) string {
	urlID := person.GetMainTreePositionID()
	if urlID == "" {
		urlID = person.GetBestID()
	}
	return "person/" + urlID + ".html"
}

func buildSearchIndex(allPersons *generations.PersonList) []SearchEntry {
	persons := allPersons.GetPersons()
	entries := make([]SearchEntry, 0, len(persons))
	for _, person := range persons {
		entries = append(entries, buildSearchEntry(person))
	}
	return entries
}

func buildSearchEntry(person generations.Person) SearchEntry {
	name := person.GetName().FormatFull()

	fields := []SearchField{
		{Label: "Name", Value: name},
	}

	if date := person.GetBirth().Date; date != "" {
		fields = append(fields, SearchField{Label: "Geboren", Value: formatDate(date)})
	}
	if date := person.GetBaptism().Date; date != "" {
		fields = append(fields, SearchField{Label: "Getauft", Value: formatDate(date)})
	}
	if date := person.GetDeath().Date; date != "" {
		fields = append(fields, SearchField{Label: "Gestorben", Value: formatDate(date)})
	}
	if date := person.GetBurial().Date; date != "" {
		fields = append(fields, SearchField{Label: "Begraben", Value: formatDate(date)})
	}

	for _, residence := range person.GetResidences() {
		if value := formatResidence(residence); value != "" {
			fields = append(fields, SearchField{Label: "Wohnort", Value: value})
		}
	}

	for _, job := range person.GetJobs() {
		if value := formatJob(job); value != "" {
			fields = append(fields, SearchField{Label: "Beruf", Value: value})
		}
	}

	for _, attribute := range person.GetAttributes() {
		if attribute != "" {
			fields = append(fields, SearchField{Label: "Tag", Value: attribute})
		}
	}

	return SearchEntry{
		Name:   name,
		URL:    personDetailsURL(person),
		Fields: fields,
	}
}

func formatResidence(r generations.Residence) string {
	var parts []string
	if r.Address != "" {
		parts = append(parts, r.Address)
	}
	zipCity := strings.TrimSpace(strings.TrimSpace(r.ZipCode) + " " + strings.TrimSpace(r.City))
	if zipCity != "" {
		parts = append(parts, zipCity)
	}
	if r.Place != "" {
		parts = append(parts, r.Place)
	}
	if r.Country != "" {
		parts = append(parts, r.Country)
	}
	return strings.Join(parts, ", ")
}

func formatJob(j generations.Job) string {
	if j.Type == "" && j.Company == "" {
		return ""
	}
	if j.Company != "" {
		return strings.TrimSpace(j.Type + " (" + j.Company + ")")
	}
	return j.Type
}

func buildSearchIndexJSON(allPersons *generations.PersonList) (template.JS, error) {
	data, err := json.Marshal(buildSearchIndex(allPersons))
	if err != nil {
		return "", err
	}
	return template.JS(data), nil
}
