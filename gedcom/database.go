package gedcom

import "strings"

type Database struct {
	Header      *Header
	Individuals *IndividualRecordSlice
	Families    *FamilyRecordSlice
}

func NewDatabase() *Database {
	return &Database{
		Individuals: NewIndividualRecordSlice(),
		Families:    NewFamilyRecordSlice(),
	}
}

func (d *Database) GetHeader() *Header {
	if d.Header == nil {
		return &Header{}
	}
	return d.Header
}

func (d *Database) GetIndividuals() *IndividualRecordSlice {
	return d.Individuals
}

func (d *Database) AddIndividual(i *IndividualRecord) error {
	*d.Individuals = append(*d.Individuals, i)
	return nil
}

func (d *Database) GetFamilies() *FamilyRecordSlice {
	return d.Families
}

func (d *Database) AddFamily(i *FamilyRecord) error {
	*d.Families = append(*d.Families, i)
	return nil
}

func (d *Database) String() string {
	var sb strings.Builder

	// Header
	sb.WriteString(d.GetHeader().String())

	// print all individuals
	for _, i := range *d.GetIndividuals() {
		sb.WriteString(i.String())
	}
	// print all families
	for _, f := range *d.GetFamilies() {
		sb.WriteString(f.String())
	}

	// Footer
	sb.WriteString("0 TRLR")

	return sb.String()
}
