# Familybook HTML Search Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a single search field to the HTML familybook's front page (`index.html`) that live-filters persons by name, places, jobs, addresses, tags and dates, and links straight to their person page.

**Architecture:** At render time, Go builds a flat JSON search index (one entry per person, each with a list of `{label, value}` fields) and embeds it inline in `index.html` as `window.SEARCH_INDEX` — no `fetch()`, so it works when the exported book is opened straight off disk (`file://`). A new plain-JS file (`s.js`) filters that array on every keystroke and renders a results dropdown. `s.js` itself currently has a dead `<script src="s.js">` reference in `index.html` (nothing ever wrote the file) — this plan also fixes that by adding a static-asset copy step to the renderer.

**Tech Stack:** Go (`cmd/generations` package `main`), Go `html/template`, vanilla JS (no dependencies), Go tests with testify.

## Global Constraints

- Must work when `index.html` is opened directly via `file://` (double-click) — no `fetch()`/XHR on separate files.
- Matching is substring, case- and umlaut-insensitive (`munchen` matches `München`) — no fuzzy/typo-tolerant matching.
- No external JS libraries — vanilla JS only.
- The search field lives only on `index.html` (front page), not on every page.
- JSON is embedded via Go's `encoding/json`, whose default `SetEscapeHTML(true)` behavior (escapes `<`, `>`, `&`) must be relied on for safe embedding inside `<script>` — do not disable it.

---

## File Map

| File | Responsibility |
|------|----------------|
| `cmd/generations/familybook_search_index.go` | New: `SearchEntry`/`SearchField` types, `buildSearchIndex`, `buildSearchIndexJSON`, `personDetailsURL` (extracted, shared) |
| `cmd/generations/familybook_search_index_test.go` | New: unit tests for the above |
| `cmd/generations/familybook_renderer_html.go` | Modify: `personHelpers()` reuses `personDetailsURL`; `Render()` computes the search index once and reuses the existing `allPersons` list; `Render()` copies `s.js` into the output dir |
| `cmd/generations/cmd_familybook.go` | Modify: `RenderConfig` gets a `SearchIndexJSON` field |
| `cmd/generations/templates/familybook/html/src/index.html` | Modify: search `<input>` + results container + inline `window.SEARCH_INDEX` |
| `cmd/generations/templates/familybook/html/src/s.js` | New: client-side search/filter/render logic |

---

### Task 1: Go search index builder

**Files:**
- Create: `cmd/generations/familybook_search_index.go`
- Create: `cmd/generations/familybook_search_index_test.go`
- Modify: `cmd/generations/familybook_renderer_html.go:195-207` (`personHelpers()`)

**Interfaces:**
- Consumes: `generations.Person` interface (`GetName() generations.Name`, `GetBirth/GetBaptism/GetDeath/GetBurial() generations.DatePlace` with `.Date string`, `GetResidences() []generations.Residence`, `GetJobs() []generations.Job`, `GetAttributes() []string`, `GetMainTreePositionID() string`, `GetBestID() string`); `generations.Name.FormatFull() string`; `generations.PersonList.GetPersons() []generations.Person`; package-local `formatDate(input string) string` (already defined in `cmd_familybook.go`).
- Produces: `type SearchField struct { Label, Value string }`, `type SearchEntry struct { Name, URL string; Fields []SearchField }`, `func personDetailsURL(person generations.Person) string`, `func buildSearchIndex(allPersons *generations.PersonList) []SearchEntry`, `func buildSearchIndexJSON(allPersons *generations.PersonList) (template.JS, error)` — all consumed by Task 2.

- [ ] **Step 1: Write the failing tests**

Create `cmd/generations/familybook_search_index_test.go`:

```go
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jojomi/generations"
	"github.com/stretchr/testify/assert"
)

func searchIndexTestDatabase(t *testing.T) *generations.MemoryDatabase {
	t.Helper()
	data := `persons:
- id: full
  gender: male
  name:
    first: ["Peter"]
    last: Stahl
  birth:
    date: "1643-03-12"
  death:
    date: "1700-05-20"
  residences:
    - address: "Hauptstr. 5"
      zip-code: "74670"
      city: "Forchtenberg"
  jobs:
    - type: "Bauer"
      company: "Hof Stahl"
  attributes:
    - "Vollspitzahn"
- id: minimal
  gender: female
  name:
    first: ["Anna"]
    last: Muster
`
	filename := filepath.Join(t.TempDir(), "search-index.yml")
	err := os.WriteFile(filename, []byte(data), 0644)
	assert.Nil(t, err)

	database := generations.NewMemoryDatabase()
	err = database.ParseYamlFile(filename)
	assert.Nil(t, err)
	return database
}

func TestBuildSearchIndexFullPerson(t *testing.T) {
	database := searchIndexTestDatabase(t)
	person, err := database.GetByID("full")
	assert.Nil(t, err)

	list := generations.NewPersonList()
	list.AddSlice([]generations.Person{person})

	entries := buildSearchIndex(&list)
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "Peter Stahl", entry.Name)
	assert.Equal(t, "person/full.html", entry.URL)
	assert.Equal(t, []SearchField{
		{Label: "Name", Value: "Peter Stahl"},
		{Label: "Geboren", Value: "12.03.1643"},
		{Label: "Gestorben", Value: "20.05.1700"},
		{Label: "Wohnort", Value: "Hauptstr. 5, 74670 Forchtenberg"},
		{Label: "Beruf", Value: "Bauer (Hof Stahl)"},
		{Label: "Tag", Value: "Vollspitzahn"},
	}, entry.Fields)
}

func TestBuildSearchIndexMinimalPerson(t *testing.T) {
	database := searchIndexTestDatabase(t)
	person, err := database.GetByID("minimal")
	assert.Nil(t, err)

	list := generations.NewPersonList()
	list.AddSlice([]generations.Person{person})

	entries := buildSearchIndex(&list)
	assert.Len(t, entries, 1)

	entry := entries[0]
	assert.Equal(t, "Anna Muster", entry.Name)
	assert.Equal(t, "person/minimal.html", entry.URL)
	assert.Equal(t, []SearchField{
		{Label: "Name", Value: "Anna Muster"},
	}, entry.Fields)
}

func TestBuildSearchIndexEmptyList(t *testing.T) {
	list := generations.NewPersonList()
	entries := buildSearchIndex(&list)
	assert.Equal(t, []SearchEntry{}, entries)
}

func TestBuildSearchIndexJSON(t *testing.T) {
	database := searchIndexTestDatabase(t)
	person, err := database.GetByID("minimal")
	assert.Nil(t, err)

	list := generations.NewPersonList()
	list.AddSlice([]generations.Person{person})

	js, err := buildSearchIndexJSON(&list)
	assert.Nil(t, err)

	var decoded []SearchEntry
	err = json.Unmarshal([]byte(js), &decoded)
	assert.Nil(t, err)
	assert.Equal(t, buildSearchIndex(&list), decoded)
}
```

- [ ] **Step 2: Run tests to verify they fail**

Run: `cd /run/storage/encrypted/data/Seafile/Stammbaum/generations && go test ./cmd/generations/... -run TestBuildSearchIndex -v`
Expected: FAIL to compile — `undefined: buildSearchIndex`, `undefined: SearchField`, etc.

- [ ] **Step 3: Write `familybook_search_index.go`**

Create `cmd/generations/familybook_search_index.go`:

```go
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
```

- [ ] **Step 4: Remove the now-duplicated `getPersonDetailsURL` closure**

In `cmd/generations/familybook_renderer_html.go`, replace the `personHelpers()` function (currently lines 195-207):

```go
func personHelpers() template.FuncMap {
	return template.FuncMap{
		"getPersonDetailsURL": func(person generations.Person) string {
			urlID := ""
			if person.GetMainTreePositionID() != "" {
				urlID = person.GetMainTreePositionID()
			} else {
				urlID = person.GetBestID()
			}
			return "person/" + urlID + ".html"
		},
	}
}
```

with:

```go
func personHelpers() template.FuncMap {
	return template.FuncMap{
		"getPersonDetailsURL": personDetailsURL,
	}
}
```

- [ ] **Step 5: Run tests to verify they pass**

Run: `cd /run/storage/encrypted/data/Seafile/Stammbaum/generations && go test ./cmd/generations/... -run TestBuildSearchIndex -v`
Expected: PASS (4 tests: `TestBuildSearchIndexFullPerson`, `TestBuildSearchIndexMinimalPerson`, `TestBuildSearchIndexEmptyList`, `TestBuildSearchIndexJSON`)

- [ ] **Step 6: Full package build check**

Run: `cd /run/storage/encrypted/data/Seafile/Stammbaum/generations && go build ./cmd/generations/...`
Expected: no output, exit code 0 (confirms `personHelpers()` still compiles against the shared `personDetailsURL`)

- [ ] **Step 7: Commit**

```bash
cd /run/storage/encrypted/data/Seafile/Stammbaum/generations
git add cmd/generations/familybook_search_index.go cmd/generations/familybook_search_index_test.go cmd/generations/familybook_renderer_html.go
git commit -m "feat(familybook-html): add search index builder"
```

---

### Task 2: Wire the search index into the renderer and front page

**Files:**
- Modify: `cmd/generations/cmd_familybook.go:64-73` (`RenderConfig` struct)
- Modify: `cmd/generations/familybook_renderer_html.go:34-100` (`Render()`)
- Modify: `cmd/generations/templates/familybook/html/src/index.html:10-15`

**Interfaces:**
- Consumes: `buildSearchIndexJSON(allPersons *generations.PersonList) (template.JS, error)` from Task 1.
- Produces: `RenderConfig.SearchIndexJSON template.JS` field, populated on every HTML render; rendered `index.html` contains `id="search-input"`, `id="search-results"`, and `window.SEARCH_INDEX` — consumed by Task 3's `s.js`.

- [ ] **Step 1: Add the field to `RenderConfig`**

In `cmd/generations/cmd_familybook.go`, `RenderConfig` currently reads (lines 64-73):

```go
type RenderConfig struct {
	Config       *FamilybookConfig
	Generations  []Generation
	ExtraPersons *generations.PersonList
	EmptyEntries *generations.PersonList
	Options      map[string]interface{}
	Stats        *generations.Stats
	Lineages     *generations.LineageContainer
	SubtreeRoots []SubtreeConfig
}
```

Add `SearchIndexJSON` (the file already imports `html/template` as `htmlTemplate`):

```go
type RenderConfig struct {
	Config          *FamilybookConfig
	Generations     []Generation
	ExtraPersons    *generations.PersonList
	EmptyEntries    *generations.PersonList
	Options         map[string]interface{}
	Stats           *generations.Stats
	Lineages        *generations.LineageContainer
	SubtreeRoots    []SubtreeConfig
	SearchIndexJSON htmlTemplate.JS
}
```

- [ ] **Step 2: Compute the index once in `Render()` and reuse the person list**

In `cmd/generations/familybook_renderer_html.go`, `Render()` currently computes `allPersons` twice under two different meanings: it's unused before the template loop, then declared with `:=` right before the person-page loop (line 71). Replace the whole block from the `// loop through different subtemplates` comment down through the `allPersons := allPersons(...)` line with:

```go
	// build the search index once and reuse the person list below
	allPersonsList := allPersons(renderConfig.Generations, renderConfig.ExtraPersons)
	searchIndexJSON, err := buildSearchIndexJSON(allPersonsList)
	if err != nil {
		return errors.Annotate(err, "building search index failed")
	}
	renderConfig.SearchIndexJSON = searchIndexJSON

	// loop through different subtemplates (index, persons, and so on)
	tplFilenames := []string{"index.html", "spitzahnen.html", "zeitleiste.html", "recherche.html"}

	for _, tplFilename := range tplFilenames {
		err = f.renderTemplateFile(tplFilename, tplFilename, makeHTMLFuncMap(toHTMLFuncMap(makeCustomFuncMap(nil)), templateHelpers(), toHTMLFuncMap(spacingHelpers()), toHTMLFuncMap(symbols()), personHelpers()), renderConfig)
		if err != nil {
			return errors.Annotatef(err, "rendering template %s failed", tplFilename)
		}
	}

	tplFilename := "person/details.html"
	workerCount := runtime.NumCPU()
	wp := workerpool.New(workerCount)
	errc := make(chan error, 1)
```

Then, still in `Render()`, the person-rendering loop (originally `for _, person := range allPersons.GetPersons() {`) and the `renderPerson(...)` call inside it both reference the old `allPersons` variable — rename both occurrences to `allPersonsList`:

```go
	for _, person := range allPersonsList.GetPersons() {
		wp.Submit(func(person generations.Person, errorChannel chan error) func() {
			return func() {
				if len(errorChannel) > 0 {
					fmt.Print("-")
					return
				}
				outputFilename := "person/" + person.GetMainTreePositionID() + ".html"
				err = f.renderPerson(tplFilename, outputFilename, renderConfig, makeHTMLFuncMap(templateHelpers(), toHTMLFuncMap(spacingHelpers()), toHTMLFuncMap(symbols()), personHelpers()), allPersonsList, person)
```

- [ ] **Step 3: Add the search UI to `index.html`**

In `cmd/generations/templates/familybook/html/src/index.html`, current lines 10-15:

```html
<h1>{{ .Config.Title }}</h1>
<div class="text-sm mb-4">
  <span class="font-bold">Stand: </span>{{ .Config.GetNow.Format .Config.DateFormat }}
</div>

{{ with .Generations }}
```

Replace with:

```html
<h1>{{ .Config.Title }}</h1>
<div class="text-sm mb-4">
  <span class="font-bold">Stand: </span>{{ .Config.GetNow.Format .Config.DateFormat }}
</div>

<div class="relative mb-4 print:hidden">
  <input type="search" id="search-input" autocomplete="off" class="w-full border border-gray-400 rounded px-3 py-2" placeholder="Suche nach Name, Ort, Beruf, Adresse, Tag, Datum …">
  <div id="search-results" class="hidden absolute z-10 w-full bg-white border border-gray-400 rounded mt-1 max-h-96 overflow-y-auto"></div>
</div>
<script>window.SEARCH_INDEX = {{ .SearchIndexJSON }};</script>

{{ with .Generations }}
```

- [ ] **Step 4: Build check**

Run: `cd /run/storage/encrypted/data/Seafile/Stammbaum/generations && go build ./cmd/generations/...`
Expected: no output, exit code 0

- [ ] **Step 5: Run the HTML familybook and verify the output**

Run: `cd /run/storage/encrypted/data/Seafile/Stammbaum/generations && .toolbox/familybook-html`
Expected: build succeeds (same output as before, plus the new markup)

Run: `grep -c 'id="search-input"' cmd/generations/output/familybook-felicitas/index.html` → expect `1`
Run: `grep -o 'window.SEARCH_INDEX = \[{"n"' cmd/generations/output/familybook-felicitas/index.html` → expect one match (confirms the array is populated, not empty/`null`)

- [ ] **Step 6: Commit**

```bash
cd /run/storage/encrypted/data/Seafile/Stammbaum/generations
git add cmd/generations/cmd_familybook.go cmd/generations/familybook_renderer_html.go cmd/generations/templates/familybook/html/src/index.html
git commit -m "feat(familybook-html): embed search index and search field on front page"
```

---

### Task 3: Client-side search (`s.js`) + static asset copy

**Files:**
- Create: `cmd/generations/templates/familybook/html/src/s.js`
- Modify: `cmd/generations/familybook_renderer_html.go:100-133` (`Render()`, just before the tailwind build step)

**Interfaces:**
- Consumes: `window.SEARCH_INDEX` (array of `{n: string, u: string, f: [{l: string, v: string}]}`) and `#search-input`/`#search-results` DOM elements from Task 2.
- Produces: `cmd/generations/output/familybook-felicitas/s.js` present in the rendered output (fixes the previously dead `<script src="s.js">` reference).

- [ ] **Step 1: Write `s.js`**

Create `cmd/generations/templates/familybook/html/src/s.js`:

```js
(function () {
  "use strict";

  function normalize(value) {
    return value
      .toLowerCase()
      .normalize("NFD")
      .replace(/[\u0300-\u036f]/g, "")
      .replace(/ß/g, "ss");
  }

  function renderResults(container, query, entries) {
    container.innerHTML = "";

    if (query === "") {
      container.classList.add("hidden");
      return;
    }

    var normalizedQuery = normalize(query);
    var matches = [];
    var seenURLs = {};

    for (var i = 0; i < entries.length; i++) {
      var entry = entries[i];
      if (seenURLs[entry.u]) {
        continue;
      }
      for (var j = 0; j < entry.f.length; j++) {
        var field = entry.f[j];
        if (normalize(field.v).indexOf(normalizedQuery) !== -1) {
          matches.push({ entry: entry, field: field });
          seenURLs[entry.u] = true;
          break;
        }
      }
      if (matches.length >= 20) {
        break;
      }
    }

    if (matches.length === 0) {
      var empty = document.createElement("div");
      empty.className = "px-3 py-2 text-gray-500";
      empty.textContent = "Keine Treffer für „" + query + "“";
      container.appendChild(empty);
      container.classList.remove("hidden");
      return;
    }

    matches.forEach(function (match) {
      var link = document.createElement("a");
      link.href = match.entry.u;
      link.className = "block px-3 py-2 hover:bg-gray-200 border-b border-gray-200";
      link.textContent = match.entry.n + " — " + match.field.v + " (" + match.field.l + ")";
      container.appendChild(link);
    });

    container.classList.remove("hidden");
  }

  document.addEventListener("DOMContentLoaded", function () {
    var input = document.getElementById("search-input");
    var results = document.getElementById("search-results");
    if (!input || !results) {
      return;
    }
    var entries = window.SEARCH_INDEX || [];

    input.addEventListener("input", function () {
      renderResults(results, input.value.trim(), entries);
    });

    input.addEventListener("keydown", function (event) {
      if (event.key === "Escape") {
        input.value = "";
        results.classList.add("hidden");
        results.innerHTML = "";
      }
    });

    document.addEventListener("click", function (event) {
      if (event.target !== input && !results.contains(event.target)) {
        results.classList.add("hidden");
      }
    });
  });
})();
```

- [ ] **Step 2: Copy `s.js` into the output directory**

In `cmd/generations/familybook_renderer_html.go`, `Render()` currently goes straight from the person-rendering loop into the tailwind build step (`// tailwind` comment, around line 102). Insert a static-asset copy step right before that comment:

```go
	// static assets
	sJSSource := sc.AbsPath(filepath.Join(f.getTemplateBasePath(), "s.js"))
	sJSData, err := os.ReadFile(sJSSource)
	if err != nil {
		return errors.Annotatef(err, "could not read %s", sJSSource)
	}
	err = os.WriteFile(filepath.Join(f.targetDir, "s.js"), sJSData, 0640)
	if err != nil {
		return errors.Annotatef(err, "could not write s.js to output directory")
	}

	// tailwind
```

- [ ] **Step 3: Build check**

Run: `cd /run/storage/encrypted/data/Seafile/Stammbaum/generations && go build ./cmd/generations/...`
Expected: no output, exit code 0

- [ ] **Step 4: Run the HTML familybook and verify `s.js` is present**

Run: `cd /run/storage/encrypted/data/Seafile/Stammbaum/generations && .toolbox/familybook-html`
Expected: build succeeds

Run: `test -s cmd/generations/output/familybook-felicitas/s.js && echo OK`
Expected: `OK` (file exists and is non-empty)

Run: `grep -c 'DOMContentLoaded' cmd/generations/output/familybook-felicitas/s.js`
Expected: `1`

- [ ] **Step 5: Commit**

```bash
cd /run/storage/encrypted/data/Seafile/Stammbaum/generations
git add cmd/generations/templates/familybook/html/src/s.js cmd/generations/familybook_renderer_html.go
git commit -m "feat(familybook-html): client-side search filtering, ship s.js"
```

---

### Task 4: End-to-end browser verification

**Files:** none (verification only — no JS test framework exists in this project, matching its established convention of manual browser verification for HTML/JS output).

**Interfaces:**
- Consumes: the fully wired feature from Tasks 1-3.
- Produces: nothing — this task is the acceptance gate for the feature.

- [ ] **Step 1: Rebuild**

Run: `cd /run/storage/encrypted/data/Seafile/Stammbaum/generations && .toolbox/familybook-html`

- [ ] **Step 2: Open the output directly off disk**

Open `cmd/generations/output/familybook-felicitas/index.html` in a browser via its `file://` path — **do not** serve it through a local HTTP server, since that's the actual constraint this feature has to satisfy.

- [ ] **Step 3: Verify each search dimension produces at least one result**

Type a query that should match, in turn: a person's name; a place name typed without umlauts (e.g. `munchen` for `München`, if such a place exists in the data — otherwise any place substring in lowercase); a job/profession keyword; part of an address; a tag/attribute value; a year (e.g. `1900`). For each: confirm the dropdown shows results and each result line shows "Name — matched value (Label)".

- [ ] **Step 4: Verify the no-match and empty states**

Type a nonsense query (e.g. `zzzzzz`) → dropdown shows "Keine Treffer für „zzzzzz“". Clear the field → dropdown hides.

- [ ] **Step 5: Verify navigation and dismissal**

Click a result → confirm it navigates to the correct person page (still via `file://`, no broken links). Re-open the search, then press Escape → dropdown closes and field clears. Re-open the search, then click outside the input/dropdown → dropdown closes.

- [ ] **Step 6: Report result**

If all checks in Steps 3-5 pass, the feature is complete — no commit needed for this task (verification only). If anything fails, return to the relevant task (1-3) and fix before proceeding.
