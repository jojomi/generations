package generations

type FlatRelationship struct {
	PartnerID  string    `yaml:"partner_id,omitempty"`
	Engagement DatePlace `yaml:"engagement,omitempty"`
	Marriage   DatePlace `yaml:"marriage,omitempty"`
	Divorce    DatePlace `yaml:"divorce,omitempty"`

	Person FlatPerson `yaml:"-"`
}

func (f FlatRelationship) GetPartner() (Person, error) {
	partner, err := f.Person.Database.GetByID(f.PartnerID)
	if err != nil {
		return nil, err
	}
	return partner, nil
}

func (f FlatRelationship) GetEngagement() *DatePlace {
	return &f.Engagement
}

func (f FlatRelationship) GetMarriage() *DatePlace {
	return &f.Marriage
}

func (f FlatRelationship) GetDivorce() *DatePlace {
	return &f.Divorce
}
