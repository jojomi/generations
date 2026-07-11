package generations

// Source is the definition of a material or digital source
type Source struct {
	ID          string `yaml:",omitempty"`
	Description string `yaml:",omitempty"`
	PDF         string `yaml:",omitempty"`
	Pages       string `yaml:",omitempty"`
	Rotate      int    `yaml:",omitempty"`
}

// Reference type
type Reference struct {
	ID          string `yaml:",omitempty"`
	Description string `yaml:",omitempty"`
}

func (s *Reference) String() string {
	if s.Description == "" {
		return s.ID
	}
	return s.Description
}

func (s *Reference) Ref() string {
	return s.ID
}

type ReferenceList []Reference
