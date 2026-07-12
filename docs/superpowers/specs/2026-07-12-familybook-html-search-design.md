# Design: Suche im Familybook (HTML-Export)

## Ziel
Auf der Startseite (`index.html`) der HTML-Ausgabe des Familybooks gibt es ein einzelnes Suchfeld. Eingaben durchsuchen live Name, Orte (Wohnorte/Geburts-/Sterbeort), Berufe, Adressen, Tags (Attributes) und Datumsangaben aller Personen. Treffer erscheinen als Dropdown unter dem Feld; Klick führt direkt zur Personenseite.

## Scope
- Nur HTML-Ausgabe (`FamilybookRendererHTML`). PDF-Ausgabe nicht betroffen.
- Suchfeld nur auf der Startseite (`index.html`), nicht auf jeder Unterseite.
- Muss funktionieren, wenn die Ausgabedateien lokal per Doppelklick geöffnet werden (`file://`) — kein `fetch()`/XHR auf externe Dateien, da das unter Chrome per `file://` blockiert wird.
- Matching: Teilstring, Groß-/Kleinschreibung und Umlaute normalisiert (`München` matcht `munchen`), kein Fuzzy-Matching.
- Ergebnisdarstellung: Name + der konkrete getroffene Wert samt Label (z. B. „Peter Stahl — Oberkessach (Wohnort)“).
- Kein Server, keine externe JS-Library — reines Vanilla JS.

## Architektur

### 1. Go: Suchindex bauen
**Neue Datei:** `cmd/generations/familybook_search_index.go`

```go
type SearchField struct {
    Label string `json:"l"`
    Value string `json:"v"`
}

type SearchEntry struct {
    Name   string        `json:"n"`
    URL    string        `json:"u"`
    Fields []SearchField `json:"f"`
}

func buildSearchIndex(allPersons *generations.PersonList) []SearchEntry
```

**Pro Person werden folgende Felder aufgenommen (jeweils nur wenn nicht leer):**
- Name: `Label "Name"`, `Value` = vollständiger Anzeigename (wie `personLinkGendered.html` ihn zeigt)
- Geburt/Taufe/Tod/Begräbnis: je ein Field mit Label `Geboren`/`Getauft`/`Gestorben`/`Begraben`, `Value` = `formatDate(...)` (gleiche Formatierung wie auf der Personenseite — deckt volle Datumssuche wie auch reine Jahressuche ab, da das Jahr Teil des formatierten Strings ist)
- Wohnorte: pro `Residence` ein Field, Label `Wohnort`, `Value` = zusammengesetzt aus Address/ZipCode/City/Place (leere Teile übersprungen)
- Berufe: pro `Job` ein Field, Label `Beruf`, `Value` = Type (+ Company falls vorhanden)
- Tags: pro Eintrag in `GetAttributes()` ein Field, Label `Tag`, `Value` = Attributwert

`URL` = `person/<MainTreePositionID>.html` (gleiche Logik wie `getPersonDetailsURL`).

**Edge Cases:**
- Person ganz ohne Zusatzdaten → nur `Name`-Field, `Fields` sonst leer. Kommt trotzdem in den Index (Name-Suche muss immer funktionieren).
- Keine Personen → leeres Array `[]`.
- Bestehende `hide-*`-Optionen (`hide-image`, `hide-lineage`, `hide-biography`) betreffen keines der gesuchten Felder → keine Wechselwirkung, Suchindex ignoriert sie.

### 2. Einbettung ins Template
**Datei:** `familybook_renderer_html.go`

- Vor dem Rendern von `index.html`: `buildSearchIndex(allPersons)` aufrufen, mit `encoding/json` marshalen, als `template.JS` in die Daten für `index.html` einhängen (neues Feld, nur für diesen einen Template-Aufruf gesetzt — nicht in das allgemeine `RenderConfig`, sondern über eine kleine Wrapper-Struct für diesen einen Aufruf).
- `encoding/json` escaped `<`, `>`, `&` standardmäßig (`SetEscapeHTML` ist per Default `true`) — das JSON ist damit sicher in einem `<script>`-Tag einbettbar, ohne dass Personendaten (z. B. ein Kommentar mit `</script>`) das Tag aufbrechen können.

**Datei:** `templates/familybook/html/src/index.html`
- Neues Suchfeld + Ergebnis-Container oberhalb der Generationen-Liste:
  ```html
  <div class="relative mb-4">
    <input type="search" id="search-input" class="w-full" placeholder="Suche nach Name, Ort, Beruf, Adresse, Tag, Datum …">
    <div id="search-results" class="absolute hidden ..."></div>
  </div>
  <script>window.SEARCH_INDEX = {{ .SearchIndexJSON }};</script>
  ```
- `<script src="s.js">` existiert im Template bereits, zeigt aber aktuell ins Leere (`s.js` wird nirgends erzeugt) — wird im Zuge dieses Features erstmals real ausgeliefert (siehe Punkt 4).

### 3. Client-seitige Suche
**Neue Datei:** `templates/familybook/html/src/s.js`

- Vanilla JS, keine Abhängigkeiten.
- Normalisierung: lowercase, Unicode-`NFD`-Normalisierung + Entfernen kombinierender Zeichen (löst Umlaute korrekt auf: `ü`→`u`, `ö`→`o`, `ä`→`a`), zusätzlich `ß`→`ss` explizit (wird durch NFD nicht aufgelöst).
- Bei jedem `input`-Event: Query normalisieren, alle `Fields` aller `SEARCH_INDEX`-Einträge auf Teilstring-Treffer prüfen (Name selbst zählt implizit auch als Field), Treffer sammeln, auf ca. 20 begrenzen.
- Pro Treffer eine Ergebniszeile: Link (`<a href="{url}">`) mit Personenname + getroffenem Wert + Label. Eine Person kann mehrfach auftreten, wenn mehrere Fields matchen — für v1 wird nur der erste Treffer pro Person gezeigt (Deduplizierung nach `URL`).
- Kein Treffer → Hinweistext „Keine Treffer für „…““.
- Dropdown schließt bei Escape, bei Klick außerhalb, bei leerem Suchfeld.
- Kein Debounce nötig (Datensatz ist klein genug, `input`-Event direkt verarbeitet).

### 4. Statische Datei ausliefern
**Datei:** `familybook_renderer_html.go`, Funktion `Render()`

- Aktuell gibt es **keinen** generischen Copy-Schritt für statische Assets (`s.js` wird nirgends geschrieben — der bestehende `<script src="s.js">`-Tag ist aktuell ein toter 404-Link). Neuer Schritt: `s.js` aus `getTemplateBasePath()/s.js` nach `f.targetDir/s.js` kopieren (einfacher Datei-Copy, kein Templating nötig, da reines JS ohne Platzhalter).

## Datenfluss
1. `Render()` sammelt `allPersons` (bereits vorhanden).
2. `buildSearchIndex(allPersons)` → `[]SearchEntry`.
3. JSON-Marshal → `template.JS`-Wert → Teil der Daten für den `index.html`-Render-Aufruf.
4. `index.html` wird geschrieben, enthält `window.SEARCH_INDEX` inline sowie `<script src="s.js">`.
5. `s.js` wird 1:1 in den Output-Ordner kopiert.
6. Browser (auch offline via `file://`): Nutzer tippt im Suchfeld → `s.js` filtert `window.SEARCH_INDEX` → Dropdown zeigt Treffer → Klick navigiert zu `person/<id>.html` (normaler relativer Link, funktioniert ohne Server).

## Tests
- **Go-Unit-Test** (`familybook_search_index_test.go`): `buildSearchIndex` mit Fixture-Personen (Person ohne Zusatzdaten, Person mit mehreren Residences/Jobs/Tags, Person mit unvollständigem Datum) → erwartete `Fields`/`Label`/`Value`/`URL` prüfen. Leere `PersonList` → leeres Ergebnis.
- **Manuell im Browser** (kein JS-Testframework im Projekt vorhanden, projektüblich manuelle Verifikation): `.toolbox/familybook-html` laufen lassen, `index.html` **direkt per Doppelklick/`file://`** öffnen (nicht über einen lokalen Server — das ist der eigentliche Härtetest), Suche mit Umlaut-Query prüfen, Klick auf Treffer prüfen, Leerstring/Keine-Treffer-Zustand prüfen, Dropdown-Schließen (Escape, Klick außerhalb) prüfen.

## Nicht im Scope
- Suche von Unterseiten aus (nur Startseite hat das Suchfeld).
- Fuzzy-/Typo-tolerante Suche.
- Volltextsuche über Biographie-/Kommentartexte.
- Index (Namensregister), Kalender, Statistik-Seite, Subtrees/Fancharts für HTML-Export (separate Specs).
- Bilder im HTML-Export (`images/`-Subdir) — separates, kleines Folge-Spec.
