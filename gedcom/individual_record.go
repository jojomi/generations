package gedcom

import "github.com/jojomi/strtpl"

// p. 61
type IndividualRecord struct {
	XREF               string
	Sex                string
	Name               *PersonalNameStructure
	Events             *IndividualEventStructure
	ChildToFamilyLink  *ChildToFamilyLink
	SpouseToFamilyLink []*SpouseToFamilyLink

	// non-gedcom?
	ChildNumber int
}

func NewIndividualRecord() *IndividualRecord {
	return &IndividualRecord{
		Events:            &IndividualEventStructure{},
		ChildToFamilyLink: &ChildToFamilyLink{},
	}
}

func (i *IndividualRecord) GetXREF() string {
	return i.XREF
}

func (i *IndividualRecord) SetXREF(xref string) *IndividualRecord {
	i.XREF = xref
	return i
}

func (i *IndividualRecord) GetSex() string {
	return i.Sex
}

func (i *IndividualRecord) SetSex(sex string) *IndividualRecord {
	i.Sex = sex
	return i
}

func (i *IndividualRecord) GetName() *PersonalNameStructure {
	return i.Name
}

func (i *IndividualRecord) SetName(name *PersonalNameStructure) *IndividualRecord {
	i.Name = name
	return i
}

func (i *IndividualRecord) String() string {
	b := NewBlock()
	b.Add(strtpl.MustEval(`{{ with .GetXREF }}@{{ . }}@ {{ end -}} INDI`, i))
	if i.Sex != "" {
		b.AddDown(strtpl.MustEval(`SEX {{ .GetSex }}`, i)).Up()
	}
	if i.Name != nil {
		b.AddDown(strtpl.MustEval(`NAME {{ .GetName.PersonalName }}`, i))
	}
	if v := i.Events.Birth.StringForLevel(1); v != "" {
		b.AddRawNewLine(v)
	}
	if v := i.Events.Death.StringForLevel(1); v != "" {
		b.AddRawNewLine(v)
	}
	b.Reset()
	if xref := i.ChildToFamilyLink.XREF; xref != "" {
		b.AddDown(strtpl.MustEval(`FAMC @{{ . }}@`, xref)).Up()
	}
	for _, s := range i.SpouseToFamilyLink {
		b.AddDown(strtpl.MustEval(`FAMS @{{ .XREF }}@`, s)).Up()
	}
	return b.String()
}

type IndividualRecordSlice []*IndividualRecord

func NewIndividualRecordSlice() *IndividualRecordSlice {
	return &IndividualRecordSlice{}
}

type SpouseToFamilyLink struct {
	XREF string // xref:fam
}
