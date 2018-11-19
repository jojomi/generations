package generations

//go:generate go-enum -f=render_person_options.go

/* ENUM(
Birth = 1
Current
CurrentAndBirth
*/
type LastnamePolicy int

type RenderPersonOptions struct {
	NodeType         NodeType
	TemplateFilename string

	// Display filter
	HideRootNodeHighlighting bool           `yaml:"hide-root-node-highlighting,omitempty"`
	LastnamePolicy           LastnamePolicy `yaml:"lastname-policy,omitempty"`

	// Output filter
	HideID         bool     `yaml:"hide-id,omitempty"`
	HideAttributes []string `yaml:"hide-attributes,omitempty"`
	HideGender     bool     `yaml:"hide-gender,omitempty"`
	HideName       bool     `yaml:"hide-name,omitempty"`
	HideBirth      bool     `yaml:"hide-birth,omitempty"`
	HideBaptism    bool     `yaml:"hide-baptism,omitempty"`
	HideDeath      bool     `yaml:"hide-death,omitempty"`
	HideDeathAge   bool     `yaml:"hide-death-age,omitempty"`
	HideBurial     bool     `yaml:"hide-burial,omitempty"`
	HideImage      bool     `yaml:"hide-image,omitempty"`
	HideJobs       bool     `yaml:"hide-jobs,omitempty"`
	HideFloruit    bool     `yaml:"hide-floruit,omitempty"`
	HideComment    bool     `yaml:"hide-comment,omitempty"`
	HideEngagement bool     `yaml:"hide-engagement,omitempty"`
	HideMarriage   bool     `yaml:"hide-marriage,omitempty"`
	HideDivorce    bool     `yaml:"hide-divorce,omitempty"`

	// special filters
	HidePlaces      bool `yaml:"hide-places,omitempty"`
	HideMiddleNames bool `yaml:"hide-middle-names,omitempty"`
}

func (o *RenderPersonOptions) SetDefaults() *RenderPersonOptions {
	if o.LastnamePolicy == 0 {
		o.LastnamePolicy = LastnamePolicyCurrentAndBirth
	}
	return o
}

func (o *RenderPersonOptions) HideAllData() {
	o.HideAttributes = []string{"all"}
	o.HideGender = true
	o.HideName = true
	o.HideBirth = true
	o.HideBaptism = true
	o.HideImage = true
	o.HideDeath = true
	o.HideBurial = true
	o.HideJobs = true
	o.HideFloruit = true
	o.HideComment = true
	o.HideEngagement = true
	o.HideMarriage = true
	o.HideDivorce = true
}

func (o *RenderPersonOptions) HideImageByLevel(treeOptions RenderTreeOptions, currentLevel int) *RenderPersonOptions {
	if (treeOptions.MinImageLevel == nil || *treeOptions.MinImageLevel <= currentLevel) &&
		(treeOptions.MaxImageLevel == nil || *treeOptions.MaxImageLevel >= currentLevel) {
		return o
	}
	o.HideImage = true
	return o
}

/* ENUM(
g = 1
p
c
*/
type NodeType int
