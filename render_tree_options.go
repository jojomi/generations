package generations

//go:generate go-enum -f=render_tree_options.go

/* ENUM(
parent = 1
child
sandclock
*/
type GraphType int

/* ENUM(
MaleFirst = 1
FemaleFirst
*/
type GenderOrder int

const GenerationsNone = -1

type RenderTreeOptions struct {
	RenderPersonOptions *RenderPersonOptions `yaml:"render-person-options,omitempty"`

	TemplateFilenameTree               string `yaml:"template-filename-tree,omitempty"`
	TemplateFilenamePerson             string `yaml:"template-filename-person,omitempty"`
	TemplateFilenameParentTree         string `yaml:"template-filename-parent-tree,omitempty"`
	TemplateFilenameParentTreeHeadless string `yaml:"template-filename-parent-tree-headless,omitempty"`
	TemplateFilenameChildTree          string `yaml:"template-filename-child-tree,omitempty"`
	TemplateFilenameUnionTree          string `yaml:"template-filename-union-tree,omitempty"`

	GraphType GraphType

	GenderOrder GenderOrder

	// Limits
	IgnoreIDs                    []string `yaml:"ignore,omitempty"`
	MaxParentGenerations         int      `yaml:"max-parent-generations,omitempty"`
	MaxParentSiblingsGenerations int      `yaml:"max-parent-siblings-generations,omitempty"`
	MaxChildGenerations          int      `yaml:"max-child-generations,omitempty"`
	MaxChildPartnersGenerations  int      `yaml:"max-child-partners-generations,omitempty"`

	HideFamilyIDs bool `yaml:"-"`
}

func (o *RenderTreeOptions) SetDefaults() *RenderTreeOptions {
	// default generation limits
	maxGenerations := 1000
	if o.MaxParentGenerations == 0 {
		o.MaxParentGenerations = maxGenerations
	}
	if o.MaxParentSiblingsGenerations == 0 {
		o.MaxParentSiblingsGenerations = maxGenerations
	}
	if o.MaxChildGenerations == 0 {
		o.MaxChildGenerations = maxGenerations
	}
	if o.MaxChildPartnersGenerations == 0 {
		o.MaxChildPartnersGenerations = maxGenerations
	}

	// other defaults
	if o.GenderOrder == 0 {
		o.GenderOrder = GenderOrderMaleFirst
	}

	// RenderPersonOptions need to be initialized too
	if o.RenderPersonOptions == nil {
		o.RenderPersonOptions = &RenderPersonOptions{
			TemplateFilename: o.TemplateFilenamePerson,
		}
	} else {
		o.RenderPersonOptions = o.RenderPersonOptions.SetDefaults()
	}

	return o
}
