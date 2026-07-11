package generations

type SiblingList struct {
	Common  *PersonList
	WithMom []PartnerChildren
	WithDad []PartnerChildren
}

type PartnerChildren struct {
	Partner  Person
	Children *PersonList
}
