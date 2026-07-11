package generations

import (
	"crypto/sha1"
	"encoding/hex"
	"slices"
	"strings"
	"time"

	age "github.com/bearbin/go-age"
	"github.com/juju/errors"
	"github.com/rs/zerolog/log"
)

type FlatPerson struct {
	Dummy             bool               `yaml:"-"`
	ID                string             `yaml:",omitempty"`
	UUID              string             `yaml:",omitempty"`
	ChildNumber       *uint              `yaml:"child_number,omitempty"`
	RelationsCount    *uint              `yaml:"relations_count,omitempty"`
	Name              Name               `yaml:",omitempty"`
	Gender            string             `yaml:",omitempty"`
	Birth             DatePlace          `yaml:",omitempty"`
	Baptism           DatePlace          `yaml:",omitempty"`
	SaintsDay         SaintsDay          `yaml:"saintsday,omitempty"`
	Death             DatePlace          `yaml:",omitempty"`
	Dead              *bool              `yaml:",omitempty"`
	Burial            DatePlace          `yaml:",omitempty"`
	Mom               string             `yaml:",omitempty"`
	Dad               string             `yaml:",omitempty"`
	Partners          []FlatRelationship `yaml:",omitempty"`
	Attributes        []string           `yaml:",omitempty"`
	ImageFilename     string             `yaml:"image,omitempty"`
	Residences        []Residence        `yaml:",omitempty"`
	Jobs              []Job              `yaml:",omitempty"`
	Comments          []Comment          `yaml:",omitempty"`
	BiographyElements []BiographyElement `yaml:"biography,omitempty"`

	Sources       ReferenceList `yaml:"sources,omitempty"`
	GenderSources ReferenceList `yaml:"gender-sources,omitempty"`

	Database   *MemoryDatabase `yaml:"-"`
	stopPerson bool            `yaml:"-"`

	treePositionIDs []string
}

type Job struct {
	Type    string `yaml:",omitempty"`
	Company string `yaml:",omitempty"`
	Place   string `yaml:",omitempty"`
	From    string `yaml:",omitempty"`
	To      string `yaml:",omitempty"`

	DateSources ReferenceList `yaml:"date-sources,omitempty"`
	Sources     ReferenceList `yaml:",omitempty"`
}

type Comment struct {
	Text    string        `yaml:",omitempty"`
	Sources ReferenceList `yaml:",omitempty"`
}

type BiographyElement struct {
	Text    string        `yaml:",omitempty"`
	Sources ReferenceList `yaml:",omitempty"`
}

type Residence struct {
	CO      string `yaml:"co,omitempty"`
	Address string `yaml:",omitempty"`
	ZipCode string `yaml:"zip-code,omitempty"`
	City    string `yaml:",omitempty"`
	Place   string `yaml:",omitempty"`
	Country string `yaml:",omitempty"`
	From    string `yaml:",omitempty"`
	To      string `yaml:",omitempty"`

	DateSources ReferenceList `yaml:"date-sources,omitempty"`
	Sources     ReferenceList `yaml:",omitempty"`
}

func NewDummyFlatPerson() *FlatPerson {
	return &FlatPerson{
		ID:    "<dummy>",
		Dummy: true,
	}
}

func (d *FlatPerson) Equals(p2 Person) bool {
	if d.GetID() == p2.GetID() && (d.GetID() != "" || p2.GetID() != "") {
		return true
	}
	if d.GetUUID() == p2.GetUUID() && (d.GetUUID() != "" || p2.GetUUID() != "") {
		return true
	}
	return false
}

func (d *FlatPerson) SetID(id string) {
	d.ID = id
}

func (d *FlatPerson) GetID() string {
	return d.ID
}

func (d *FlatPerson) GetUUID() string {
	return d.UUID
}

func (d *FlatPerson) GetBestID() string {
	if uuid := d.GetUUID(); uuid != "" {
		return uuid
	}
	if id := d.GetID(); id != "" {
		return id
	}
	// Fallback
	h := sha1.New()
	h.Write([]byte(strings.Join([]string{d.Name.GetUsedFirst(), d.Name.Last, d.Name.Birth}, ",")))
	sha := h.Sum(nil)
	shaStr := hex.EncodeToString(sha)
	return shaStr
}

func (d *FlatPerson) MatchesIDUUID(idUUIDSearches ...string) bool {
	for _, search := range idUUIDSearches {
		if search == "" {
			continue
		}
		if d.GetUUID() == search {
			return true
		}
		if d.GetID() == search {
			return true
		}
	}
	return false
}

func (d *FlatPerson) MatchesSearch(search string) bool {
	if d.MatchesIDUUID(search) {
		return true
	}
	name := d.GetName()
	if len(name.First) > 0 {
		if name.First[0]+" "+name.Last == search {
			return true
		}
	}
	return false
}

func (d *FlatPerson) GetGender() Gender {
	gender, err := ParseGender(d.Gender)
	if err != nil {
		return GenderUnknown
	}
	return gender
}

func (d *FlatPerson) GetGenderSources() ReferenceList {
	return d.GenderSources
}

func (d *FlatPerson) GetSources() ReferenceList {
	return d.Sources
}

func (d *FlatPerson) GetName() Name {
	return d.Name
}

func (d *FlatPerson) GetChildNumber() *uint {
	return d.ChildNumber
}

func (d *FlatPerson) SetChildNumber(childNumber *uint) {
	d.ChildNumber = childNumber
}

func (d *FlatPerson) GetRelationsCount() *uint {
	return d.RelationsCount
}

func (d *FlatPerson) SetRelationsCount(relationsCount *uint) {
	d.RelationsCount = relationsCount
}

func (d *FlatPerson) GetSiblings() (*SiblingList, error) {
	var common *PersonList
	result := SiblingList{
		Common: common,
	}
	mom, err := d.GetMom()
	if err != nil {
		return &result, errors.Annotate(err, "mom not found")
	}
	dad, err := d.GetDad()
	if err != nil {
		return &result, errors.Annotate(err, "dad not found")
	}
	if mom.IsDummy() || dad.IsDummy() {
		return &result, nil
	}
	common, err = mom.GetChildrenWith(dad)
	if err != nil {
		return &result, errors.Annotate(err, "GetChildrenWith failed")
	}
	common.Exclude(d)

	result.Common = common
	return &result, nil
}

func (d *FlatPerson) GetRelationships() []Relationship {
	result := make([]Relationship, len(d.Partners))
	for i := range d.Partners {
		item := d.Partners[i]
		item.Person = *d
		result[i] = &item
	}
	return result
}

func (d *FlatPerson) GetRelationshipWith(partner Person) Relationship {
	for _, item := range d.Partners {
		if !partner.MatchesSearch(item.PartnerID) {
			continue
		}
		item.Person = *d
		return &item
	}
	return nil
}

func (d *FlatPerson) GetBirth() DatePlace {
	return d.Birth
}

func (d *FlatPerson) GetBaptism() DatePlace {
	return d.Baptism
}

func (d *FlatPerson) GetSaintsDay() SaintsDay {
	return d.SaintsDay
}

func (d *FlatPerson) GetDeath() DatePlace {
	return d.Death
}

func (d *FlatPerson) GetDeadState() *bool {
	return d.Dead
}

func (d *FlatPerson) GetAge(now time.Time) int {
	birth := d.GetBirth()

	if now.IsZero() {
		return -1
	}

	if birth.Empty() || len(birth.Date) != 10 {
		return -1
	}
	birthTime, err := time.Parse("2006-01-02", birth.Date)
	if err != nil {
		return -1
	}

	return age.AgeAt(birthTime, now)
}

func (d *FlatPerson) GetDeathAge() int {
	birth := d.GetBirth()
	death := d.GetDeath()

	if birth.Empty() || len(birth.Date) != 10 {
		return -1
	}
	birthTime, err := time.Parse("2006-01-02", birth.Date)
	if err != nil {
		return -1
	}

	if death.Empty() || len(death.Date) != 10 {
		return -1
	}
	deathTime, err := time.Parse("2006-01-02", death.Date)
	if err != nil {
		return -1
	}

	return age.AgeAt(birthTime, deathTime)
}

func (d *FlatPerson) GetBurial() DatePlace {
	return d.Burial
}

func (d *FlatPerson) GetRawMom() string {
	return d.Mom
}

func (d *FlatPerson) GetMom() (Person, error) {
	if d.Mom == "" {
		return NewDummyFlatPerson(), nil
	}
	mom, err := d.Database.GetByID(d.Mom)
	if err != nil {
		return NewDummyFlatPerson(), err
	}
	return mom, nil
}

func (d *FlatPerson) SetRawMom(mom string) {
	d.Mom = mom
}

func (d *FlatPerson) GetRawDad() string {
	return d.Dad
}

func (d *FlatPerson) GetDad() (Person, error) {
	log.Trace().Str("person", d.String()).Str("raw", d.Dad).Msg("searching dad")
	if d.Dad == "" {
		return NewDummyFlatPerson(), nil
	}
	dad, err := d.Database.GetByID(d.Dad)
	if err != nil {
		log.Trace().Err(err).Msg("missing dad in database, open link!?")
		return NewDummyFlatPerson(), err
	}
	return dad, nil
}

func (d *FlatPerson) SetRawDad(dad string) {
	d.Dad = dad
}

func (d *FlatPerson) SetStopPerson(value bool) {
	d.stopPerson = value
}

func (d *FlatPerson) IsStopPerson() bool {
	return d.stopPerson
}

func (d *FlatPerson) GetPartners() (*PersonList, error) {
	result := NewPersonList()

	// find explicit partners
	var err error
	for _, partner := range d.Partners {
		var (
			theOther Person
		)
		partnerID := partner.PartnerID
		if partnerID == "" {
			theOther = NewDummyFlatPerson()
		} else {
			theOther, err = d.Database.Get(partnerID)
			if err != nil {
				return &result, errors.Annotatef(err, "Searching partner for person %s", d)
			}
		}
		result.AddPerson(theOther)
	}

	// find partners through common children
	childrenPartners, err := d.GetChildrenParents()
	if err != nil {
		return &result, err
	}

	// merge results
	resultPointer := &result
	finalResult := childrenPartners.AddList(resultPointer).RemoveDuplicatesByBestID()

	// sort by given index, fallback to marriage date
	finalResult.SortPersons(func(a, b Person) bool {
		aIndex := d.getRelationshipIndex(a)
		bIndex := d.getRelationshipIndex(b)

		if aIndex != bIndex {
			return aIndex < bIndex
		}

		// fallback
		aPartners := a.GetRelationships()
		if len(aPartners) == 0 {
			return false
		}
		bPartners := b.GetRelationships()
		if len(bPartners) == 0 {
			return true
		}
		return aPartners[0].GetMarriage().Date < bPartners[0].GetMarriage().Date
	})

	return finalResult, nil
}

func (d *FlatPerson) getRelationshipIndex(p Person) int {
	for _, r := range d.GetRelationships() {
		partner, err := r.GetPartner()
		if err != nil {
			continue
		}
		if partner.GetBestID() == p.GetBestID() {
			return r.GetIndex()
		}
	}
	return 0
}

func (d *FlatPerson) GetChildrenParents() (*PersonList, error) {
	result := NewPersonList()
	parentsSeen := make(map[string]struct{}, 0)
	var (
		mom       Person
		dad       Person
		candidate Person
		ok        bool
		err       error
	)
	children, err := d.GetChildren()
	if err != nil {
		return &result, err
	}
	for _, child := range children.GetPersons() {
		mom, err = child.GetMom()
		if err != nil {
			return &result, err
		}
		dad, err = child.GetDad()
		if err != nil {
			return &result, err
		}

		// if this FlatPerson is mom or dad the other one is a candidate parent to be returned
		if mom != nil && mom.MatchesIDUUID(d.GetUUID(), d.GetID()) {
			candidate = dad
		}
		if dad != nil && dad.MatchesIDUUID(d.GetUUID(), d.GetID()) {
			candidate = mom
		}

		// if the candidate has already been seen, don't add it to the parent list again
		if candidate == nil {
			candidate = NewDummyFlatPerson()
		}

		if _, ok = parentsSeen[candidate.GetID()]; ok {
			continue
		}
		parentsSeen[candidate.GetID()] = struct{}{}

		result.AddPerson(candidate)
	}
	return &result, nil
}

func (d *FlatPerson) HasChildrenCount() bool {
	return true
}

func (d *FlatPerson) GetChildrenCount() int {
	// TODO this should be a *different* value:
	// We might know a person had 5 children, but no data about them
	children, _ := d.GetChildren()
	return children.Count()
}

func (d *FlatPerson) IsEmpty() bool {
	return d.GetID() == "" && d.GetUUID() == ""
}

func (d *FlatPerson) GetChildren() (*PersonList, error) {
	result := NewPersonList()

	var (
		mom Person
		dad Person
		err error
	)
	if d.IsDummy() || d.IsEmpty() {
		return &result, nil
	}
	for _, child := range d.Database.Persons {
		mom, err = child.GetMom()
		if err != nil {
			return &result, err
		}
		if mom != nil && mom.MatchesIDUUID(d.GetUUID(), d.GetID()) {
			result.AddPerson(child)
			continue
		}
		dad, err = child.GetDad()
		if err != nil {
			return &result, err
		}
		if dad != nil && dad.MatchesIDUUID(d.GetUUID(), d.GetID()) {
			result.AddPerson(child)
			continue
		}
	}

	// just to be safe
	result.RemoveDuplicatesByBestID()

	return &result, nil
}

func (d *FlatPerson) GetChildrenWith(partner Person) (*PersonList, error) {
	result := NewPersonList()
	children, err := d.GetChildren()
	if err != nil {
		return &result, err
	}
	for _, child := range children.GetPersons() {
		mom, err := child.GetMom()
		if err != nil {
			continue
		}
		dad, err := child.GetDad()
		if err != nil {
			continue
		}
		otherParent := getOtherPerson(mom, dad, d)
		if otherParent == nil {
			continue
		}

		if partner.IsDummy() {
			if otherParent.IsDummy() {
				result.AddPerson(child)
			}
			continue
		}

		if otherParent.IsDummy() {
			continue
		}
		if otherParent.GetID() == partner.GetID() {
			result.AddPerson(child)
		}
	}

	// sort children: first by child-number, then by date of birth
	result.SortPersons(func(a, b Person) bool {
		aNum := a.GetChildNumber()
		bNum := b.GetChildNumber()
		if aNum == nil && bNum != nil {
			return true
		}
		if bNum == nil && aNum != nil {
			return false
		}
		if aNum != bNum {
			return *aNum < *bNum
		}
		return a.GetBirth().Date < b.GetBirth().Date
	})

	return &result, nil
}

func (d *FlatPerson) GetAttributes() []string {
	return d.Attributes
}

func (d *FlatPerson) AddAttribute(attr string) {
	d.Attributes = append(d.Attributes, attr)
}

func (d *FlatPerson) GetImageFilename() string {
	return d.ImageFilename
}

func (d *FlatPerson) SetImageFilename(filename string) {
	d.ImageFilename = filename
}

func (d *FlatPerson) GetJobs() []Job {
	return d.Jobs
}

func (d *FlatPerson) SetJobs(jobs []Job) {
	d.Jobs = jobs
}

func (d *FlatPerson) AddJob(job Job) {
	if d.Jobs == nil {
		d.Jobs = make([]Job, 0, 1)
	}
	d.Jobs = append(d.Jobs, job)
}

func (d *FlatPerson) GetResidences() []Residence {
	return d.Residences
}

func (d *FlatPerson) AddResidence(residence Residence) {
	if d.Residences == nil {
		d.Residences = make([]Residence, 0, 1)
	}
	d.Residences = append(d.Residences, residence)
}

func (d *FlatPerson) SetResidences(residences []Residence) {
	d.Residences = residences
}

func (d *FlatPerson) GetBiographyElements() []BiographyElement {
	return d.BiographyElements
}

func (d *FlatPerson) SetBiographyElements(biographyElements []BiographyElement) {
	d.BiographyElements = biographyElements
}

func (d *FlatPerson) AddBiographyElement(biographyElement BiographyElement) {
	if d.BiographyElements == nil {
		d.BiographyElements = make([]BiographyElement, 0, 1)
	}
	d.BiographyElements = append(d.BiographyElements, biographyElement)
}

func (d *FlatPerson) GetComments() []Comment {
	return d.Comments
}

func (d *FlatPerson) AddComment(comment Comment) {
	if d.Comments == nil {
		d.Comments = make([]Comment, 0, 1)
	}
	d.Comments = append(d.Comments, comment)
}

func (d *FlatPerson) SetComments(comments []Comment) {
	d.Comments = comments
}

func (d *FlatPerson) IsDummy() bool {
	return d == nil || d.Dummy
}

func (d *FlatPerson) IsAlive(referenceTime time.Time) bool {
	return IsAlive(d, referenceTime)
}

func (d *FlatPerson) IsSpitzahn() bool {
	if p, err := d.GetMom(); err != nil || p == nil {
		return true
	}
	if p, err := d.GetDad(); err != nil || p == nil {
		return true
	}
	return false
}

func (d *FlatPerson) GetTreePositionIDs() []string {
	return d.treePositionIDs
}

func (d *FlatPerson) GetMainTreePositionID() string {
	all := d.GetTreePositionIDs()
	if len(all) < 1 {
		return ""
	}
	return all[0]
}

func (d *FlatPerson) AddTreePositionID(ID string) Person {
	// search if already there
	if slices.Contains(d.treePositionIDs, ID) {
		return d
	}
	d.treePositionIDs = append(d.treePositionIDs, ID)
	return d
}

func (d *FlatPerson) ClearTreePositionIDs() Person {
	d.treePositionIDs = make([]string, 0)
	return d
}

func (d *FlatPerson) Copy() Person {
	newPerson := FlatPerson{
		Dummy:             d.Dummy,
		ID:                d.ID,
		UUID:              d.UUID,
		ChildNumber:       d.ChildNumber,
		Name:              d.Name,
		Gender:            d.Gender,
		Birth:             d.Birth,
		Baptism:           d.Baptism,
		Death:             d.Death,
		Dead:              d.Dead,
		Burial:            d.Burial,
		SaintsDay:         d.SaintsDay,
		Mom:               d.Mom,
		Dad:               d.Dad,
		Partners:          d.Partners,
		Attributes:        d.Attributes,
		ImageFilename:     d.ImageFilename,
		Residences:        d.Residences,
		Jobs:              d.Jobs,
		BiographyElements: d.BiographyElements,
		Comments:          d.Comments,
		Sources:           d.Sources,

		Database: d.Database,

		treePositionIDs: d.treePositionIDs,
	}

	return &newPerson
}

func (d FlatPerson) String() string {
	var b strings.Builder
	b.WriteString("/Person ")
	if d.GetID() != "" {
		b.WriteString(d.GetID())
	} else {
		b.WriteString(d.GetUUID())
	}
	b.WriteString(": ")
	if name := d.GetName(); !name.Empty() {
		for _, first := range name.First {
			b.WriteString(first)
			b.WriteString(" ")
		}
		b.WriteString(name.Last)
	}
	b.WriteString("/")
	return b.String()
}

func getOtherPerson(a, b, than Person) Person {
	if a == nil && than == nil {
		return b
	}
	if b == nil && than == nil {
		return a
	}
	if than == nil || a == nil || b == nil {
		return nil
	}
	if a.GetBestID() == than.GetBestID() {
		return b
	}
	if b.GetBestID() == than.GetBestID() {
		return a
	}
	return nil
}
