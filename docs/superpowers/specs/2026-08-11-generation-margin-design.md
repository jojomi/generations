# Generation Margin Feature — Design

Date: 2026-08-11
Status: approved, pending implementation plan

## Goal

Optional, document-wide feature. On every physical PDF page that shows
persons of exactly one generation (relative to the whole book, not the
individual tree), print that generation's German name in the outer page
margin: light gray, rotated 90° for outward-readable orientation (mirrored
between odd/even pages, as in a bound book), sized as large as possible
without overlapping the main content. Pages that mix persons from more than
one generation get no label.

"Generation" is always the absolute level (`AbsoluteLevel.Index` /
`TreeConfig.ProbandLevel`), not a per-tree relative level: e.g. the proband's
parents are level 1 ("Eltern") regardless of which tree happens to render
them.

Out of scope: `cmd/templates/tree/basic.tex` (the legacy hardcoded-level tree
template). Feature only applies to templates that loop `.Levels.Combined`
(currently `min-space.tex`, `min-space-brigitte.tex`).

## Config

New field `Config.GenerationMargin bool` (yaml: `generation-margin`),
default `false`. When `false`, nothing changes: no new packages, no new
macros, no behavior change.

Existing page margin (`geometry` `margin=1.25cm`) is **not** changed. The
label must fit inside that existing margin strip.

## Go: precise per-TreeConfig level range

`RenderGenealogytree` (package `generations`, `generator.go`) gains a second
return value describing which absolute-relative levels were *actually*
rendered (not just configured via `MaxParentGenerations` etc.), so a
TreeConfig whose configured generation caps exceed the real data depth still
gets a correct (non-suppressed) label.

```go
type LevelRange struct {
    Min, Max   int  // proband-relative: 0 = proband, positive = ancestors, negative = descendants
    HasPersons bool
}

func RenderGenealogytree(p Person, o RenderTreeOptions) ([]byte, LevelRange, error)
```

Implementation: thread a small accumulator (`*LevelRange`, with a `record(level int)`
method) through the recursive render functions in `generator_parent_tree.go`
and `generator_child_tree.go` (`renderParentTree`, `renderChildTree`,
`renderUnionData`, `renderUnion`, `renderChildrenWithPartner`), plus the
top-level sibling rendering in `RenderGenealogytree` itself. Every place that
calls `renderPerson(...)` for a real (non-dummy, non-ignored) person records
its level (ancestors: `level`; descendants: `-level`).

`cmd/main.go`, after calling `RenderGenealogytree`, converts the relative
range to absolute and stores it on the `TreeConfig`:

```go
treeConfig.MinLevel = treeConfig.ProbandLevel + levelRange.Min
treeConfig.MaxLevel = treeConfig.ProbandLevel + levelRange.Max
```

(`MinLevel`/`MaxLevel int`, `yaml:"-"`, only meaningful when `levelRange.HasPersons`.)

Both existing call sites of `RenderGenealogytree` (`commandRoot`,
`commandTest`) are updated for the new signature.

## Go: generation name table

New file `cmd/generation_names.go`:

```go
func GenerationName(level int) string {
    switch {
    case level == 0:
        return ""
    case level == 1:
        return "Eltern"
    case level == 2:
        return "Großeltern"
    case level == 3:
        return "Urgroßeltern"
    case level == 4:
        return "Alturgroßeltern"
    case level >= 5:
        return "Altvordern"
    case level == -1:
        return "Kinder"
    case level == -2:
        return "Enkel"
    case level == -3:
        return "Urenkel"
    case level == -4:
        return "Alturenkel"
    default: // level <= -5
        return "Nachfahren"
    }
}
```

Unit-tested directly (table test over representative levels incl. boundaries
0, ±4, ±5, ±20).

## LaTeX: per-page detection and rendering

### Why LaTeX-side tracking is needed at all

Go knows, per TreeConfig, whether that tree alone spans multiple
generations. It does **not** know whether two different TreeConfigs end up
sharing the same physical PDF page (page breaks are LaTeX's decision, driven
by `page-break-after` / natural content flow / `\vfill`). Cross-TreeConfig
merging can only happen at LaTeX compile time.

### Recording (once per TreeConfig)

In `min-space.tex` and `min-space-brigitte.tex`, immediately before
`\begin{tikzpicture}`, when `$.Config.GenerationMargin` is true:

```
\gtrRecordGenerationRange{ {{ .MinLevel }} }{ {{ .MaxLevel }} }
```

A tikzpicture never splits across a page break, so `\thepage` at this point
in the source always matches the physical page the tree ends up on.

### Accumulator (document preamble, plain TeX — no extra list package needed)

```
\newif\ifGenPageSeen
\newcount\GenPageMin
\newcount\GenPageMax
\newcommand{\gtrRecordGenerationRange}[2]{%
  \ifGenPageSeen
    \ifnum#1<\GenPageMin \GenPageMin=#1\relax\fi
    \ifnum#2>\GenPageMax \GenPageMax=#2\relax\fi
  \else
    \GenPageSeentrue
    \GenPageMin=#1\relax
    \GenPageMax=#2\relax
  \fi
}
```

Tracking min/max (rather than the exact set of levels) is sufficient: "page
shows more than one generation" is exactly `Min ≠ Max`, regardless of
whether the union of levels has gaps.

### Name lookup table (generated once, document preamble)

Go (in `cmd/main.go`, document-template data) emits one entry per index in
`-20..20` (matching the existing `AddDefaultLevels(-20, 20)` convention),
reusing `GenerationName`:

```
\expandafter\def\csname gtrGenName-3\endcsname{Urenkel}
\expandafter\def\csname gtrGenName4\endcsname{Alturgroßeltern}
...
```

Levels whose name is `""` (level 0) still get a (empty) definition, so the
lookup is total over the range.

### Shipout hook (new package: `atbegshi` — small, standard, no other deps)

```
\AtBeginShipout{%
  \ifGenPageSeen
    \ifnum\GenPageMin=\GenPageMax
      \edef\gtrGenCurrent{\csname gtrGenName\the\GenPageMin\endcsname}%
      \ifx\gtrGenCurrent\empty
      \else
        % draw, see below
      \fi
    \fi
  \fi
  \global\GenPageSeenfalse
}
```

Resetting `GenPageSeen` (not min/max — irrelevant once unset) after every
shipout prepares the accumulator for the next page.

### Drawing (TikZ, already loaded via `genealogytree`)

No `adjustbox`, no `eso-pic`. A `tikzpicture` overlay anchored to the
`current page` node (available immediately, no `remember picture` needed
since we don't reference named nodes across pictures):

- Anchor: `current page.south east` on odd pages, `current page.south west`
  on even pages (`\ifodd\value{page}`) — outer margin, mirrored like a bound
  book.
- Rotation: `rotate=90` / `rotate=-90` respectively, so text reads bottom-to-top
  or top-to-bottom as appropriate for outward reading.
- Color/font: `gray!35`, `\sffamily\bfseries`.
- Scale-to-fit: measure the unrotated label's natural box via
  `\settowidth`/`\settoheight` at a reference size, compute
  `scale = min(availableHeight / naturalWidth, availableMarginWidth / naturalHeight)`
  (available height ≈ `\paperheight` minus a small top/bottom pad; available
  margin width ≈ the existing 1.25cm minus a small pad so the label never
  touches the tree content), and pass that factor as the TikZ node's `scale=`
  option. This reproduces "as large as possible without distortion or
  collision" without adding a dependency.

## Testing

- Go: table test for `GenerationName` (boundaries 0, ±4, ±5, extremes).
- Go: unit tests in `generator_parent_test.go` / `generator_child_test.go`
  asserting the returned `LevelRange` for representative trees (single
  person, deep ancestor-only tree, deep descendant-only tree, mixed,
  tree cut short by missing data despite generous `MaxParentGenerations`).
- LaTeX: manual visual check (`generation-margin: true` in a test config,
  compile, inspect a same-generation page shows the label correctly
  mirrored/rotated on odd vs. even pages, and a mixed-generation page shows
  no label).

## Non-goals / explicitly deferred

- `cmd/templates/tree/basic.tex` support (hardcoded levels, not currently
  worked on).
- Configurable generation names beyond the fixed German table above.
- Configurable margin width (stays at whatever `geometry`'s `margin=` is set
  to; no dedicated reserved width).
