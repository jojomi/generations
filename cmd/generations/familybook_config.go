package main

import (
	"time"

	//"github.com/jinzhu/copier"

	"github.com/jojomi/generations"
	"github.com/ulule/deepcopier"
)

type FamilybookConfig struct {
	Databases []string `yaml:"databases,omitempty"`
	Templates struct {
		Document Template `yaml:"document,omitempty"`
		Entry    Template `yaml:"entry,omitempty"`
	} `yaml:"templates,omitempty"`

	DocumentOptions string `yaml:"document-options,omitempty"`
	Title           string `yaml:"title,omitempty"`
	Now             string `yaml:"now,omitempty"`
	DateFormat      string `yaml:"date-format,omitempty"`
	Renderer        string `yaml:"renderer,omitempty"`

	Proband      string            `yaml:"proband,omitempty"`
	ExtraPersons []string          `yaml:"extra-persons,omitempty"`
	Author       string            `yaml:"author,omitempty"`
	Attribution  string            `yaml:"attribution,omitempty"`
	PreContent   string            `yaml:"pre-content,omitempty"`
	PostContent  string            `yaml:"post-content,omitempty"`
	CustomHead   string            `yaml:"custom-head,omitempty"`
	Options      FamilybookOptions `yaml:"options,omitempty"`

	OutputFilename string `yaml:"output-filename,omitempty"`

	Sources []generations.Source `yaml:",omitempty"`
}

func (c *FamilybookConfig) Copy() *FamilybookConfig {
	var result FamilybookConfig
	deepcopier.Copy(c).To(&result)
	return &result
}

func (c *FamilybookConfig) CopyWithOptionsPersonShowEmpty() *FamilybookConfig {
	result := c.Copy()
	result.Options.Person.ShowEmpty = true
	return result
}

func (c *FamilybookConfig) GetNow() time.Time {
	now := time.Now()
	if c.Now == "" {
		return now
	}
	nowParsed, err := time.Parse("2006-01-02", c.Now)
	if err != nil {
		return now
	}
	return nowParsed
}

func (c *FamilybookConfig) GetOutputFilename() string {
	return c.OutputFilename
}

func (c *FamilybookConfig) SetOutputFilename(filename string) {
	c.OutputFilename = filename
}

func (c *FamilybookConfig) SetDefaults() {
	if c.Renderer == "" {
		c.Renderer = "pdf"
	}
	if c.Templates.Document.Filename == "" {
		c.Templates.Document.Filename = "templates/document/familybook.tex"
	}
	if c.Templates.Entry.Filename == "" {
		c.Templates.Entry.Filename = "templates/familybook/entry.tex"
	}

	c.Options.SetDefaults()
}

type FamilybookOptions struct {
	MaxParentGenerations   int  `yaml:"max-parent-generations,omitempty"`
	MaxChildGenerations    int  `yaml:"max-child-generations,omitempty"`
	IncludePartners        bool `yaml:"include-partners,omitempty"`
	IncludeIslandAsExtra   bool `yaml:"include-island-as-extra,omitempty"`
	IncludeProbandSiblings bool `yaml:"include-proband-siblings,omitempty"`
	HideSources            bool `yaml:"hide-sources,omitempty"`

	SinglePersonPages bool `yaml:"single-person-pages,omitempty"`
	ShortLists        bool `yaml:"short-lists,omitempty"`
	MultilineLists    bool `yaml:"multiline-lists,omitempty"`
	WrapNames         bool `yaml:"wrap-names,omitempty"`
	AllowLiveData     bool `yaml:"allow-live-data,omitempty"`
	AgesYearsOnly     bool `yaml:"ages-years-only,omitempty"`
	// ShowLineage       bool                   `yaml:"show-lineage,omitempty"`
	MaternalOrder  bool                   `yaml:"maternal-order,omitempty"`
	AccentColor    string                 `yaml:"accent-color,omitempty"`
	Debug          FamilybookDebugOptions `yaml:"debug,omitempty"`
	EmptyEntries   int                    `yaml:"empty-entries,omitempty"`
	Texts          FamilybookTexts        `yaml:"texts,omitempty"`
	HideSpitzahnen bool                   `yaml:"hide-spitzahnen,omitempty"`

	Subtrees FamilybookSubtreesOptions `yaml:"subtrees,omitempty"`
	Groups   FamilybookGroupsOptions   `yaml:"groups,omitempty"`
	Stats    FamilybookStatsOptions    `yaml:"stats,omitempty"`
	Calendar FamilybookCalendarOptions `yaml:"calendar,omitempty"`

	Person FamilybookPersonOptions `yaml:"person"`
}

type FamilybookDebugOptions struct {
	IDs          bool `yaml:"ids,omitempty"`
	UUIDs        bool `yaml:"uuids,omitempty"`
	Links        bool `yaml:"links,omitempty"`
	TreePos      bool `yaml:"treepos,omitempty"`
	ChildNumbers bool `yaml:"child-numbers,omitempty"`
}

type FamilybookSubtreesOptions struct {
	Hide                bool   `yaml:"hide,omitempty"`
	Type                string `yaml:"type,omitempty"` // fancharts or ahnentafel
	GenerationsPerChart int    `yaml:"generations-per-chart,omitempty"`
}

type FamilybookGroupsOptions struct {
	Hide bool `yaml:"hide,omitempty"`
}

type FamilybookStatsOptions struct {
	Hide  bool `yaml:"hide,omitempty"`
	Count int  `yaml:"count,omitempty"`

	HideNumbers              bool `yaml:"hide-numbers,omitempty"`
	HideGenerations          bool `yaml:"hide-generations,omitempty"`
	HideAlive                bool `yaml:"hide-alive,omitempty"`
	HideEarliestBirth        bool `yaml:"hide-earliest-birth,omitempty"`
	HideLatestBirth          bool `yaml:"hide-latest-birth,omitempty"`
	HideOldest               bool `yaml:"hide-oldest,omitempty"`
	HideDeathAge             bool `yaml:"hide-death-age,omitempty"`
	HideMarriageAge          bool `yaml:"hide-marriage-age,omitempty"`
	HideTotalAgeLiving       bool `yaml:"hide-total-age-living,omitempty"`
	HideLongestMarriages     bool `yaml:"hide-longest-marriages,omitempty"`
	HideLongestName          bool `yaml:"hide-longest-name,omitempty"`
	HideShortestName         bool `yaml:"hide-shortest-name,omitempty"`
	HideMostChildren         bool `yaml:"hide-most-children,omitempty"`
	HideBirthMonths          bool `yaml:"hide-birth-months,omitempty"`
	HideMostCommonLastNames  bool `yaml:"hide-most-common-last-names,omitempty"`
	HideMostCommonFirstNames bool `yaml:"hide-most-common-first-names,omitempty"`
	HideUniqueFirstNames     bool `yaml:"hide-unique-first-names,omitempty"`
	HideCommonBirthdays      bool `yaml:"hide-common-birthdays,omitempty"`
	HideDuplicateNames       bool `yaml:"hide-duplicate-names,omitempty"`
}

type FamilybookTexts struct {
	Proband string `yaml:"proband,omitempty"`
}

type FamilybookCalendarOptions struct {
	Hide      bool `yaml:"hide,omitempty"`
	HidePages bool `yaml:"hide-pages,omitempty"`
}

type FamilybookPersonOptions struct {
	ShowEmpty bool `yaml:"show-empty,omitempty"`

	HideGender               bool `yaml:"hide-gender,omitempty"`
	HideName                 bool `yaml:"hide-name,omitempty"`
	HideBirth                bool `yaml:"hide-birth,omitempty"`
	ShowAge                  bool `yaml:"show-age,omitempty"`
	HideBaptism              bool `yaml:"hide-baptism,omitempty"`
	HideDeath                bool `yaml:"hide-death,omitempty"`
	HideDeathAge             bool `yaml:"hide-death-age,omitempty"`
	HideBurial               bool `yaml:"hide-burial,omitempty"`
	HideImage                bool `yaml:"hide-image,omitempty"`
	HideJobs                 bool `yaml:"hide-jobs,omitempty"`
	HideFloruit              bool `yaml:"hide-floruit,omitempty"`
	HideComment              bool `yaml:"hide-comment,omitempty"`
	HideEngagement           bool `yaml:"hide-engagement,omitempty"`
	HideSaintsDay            bool `yaml:"hide-saints-day,omitempty"`
	HideMarriage             bool `yaml:"hide-marriage,omitempty"`
	HideMarriageAge          bool `yaml:"hide-marriage-age,omitempty"`
	HideMarriageReligious    bool `yaml:"hide-marriage-religious,omitempty"`
	HideMarriageReligiousAge bool `yaml:"hide-marriage-religious-age,omitempty"`
	HideDivorce              bool `yaml:"hide-divorce,omitempty"`
	HideBiography            bool `yaml:"hide-biography,omitempty"`
	DetailedResidences       bool `yaml:"detailed-residences,omitempty"`

	HideLineage   bool `yaml:"hide-lineage,omitempty"`
	HideLifeDates bool `yaml:"hide-life-dates,omitempty"`

	// special filters
	HidePlaces      bool `yaml:"hide-places,omitempty"`
	HideMiddleNames bool `yaml:"hide-middle-names,omitempty"`
}

func (f *FamilybookOptions) SetDefaults() {
	if f.MaxParentGenerations == 0 {
		f.MaxParentGenerations = 1000
	}
	if f.MaxChildGenerations == 0 {
		f.MaxChildGenerations = 1000
	}

	if f.Stats.Count == 0 {
		f.Stats.Count = 10
	}
}
