package generations

import (
	"time"

	"github.com/jojomi/timecalc"
	"github.com/juju/errors"
)

type FlatRelationship struct {
	PartnerID         string    `yaml:"partner_id,omitempty"`
	Index             int       `yaml:"index,omitempty"`
	Engagement        DatePlace `yaml:"engagement,omitempty"`
	Marriage          DatePlace `yaml:"marriage,omitempty"`
	MarriageReligious DatePlace `yaml:"marriage_religious,omitempty"`
	Divorce           DatePlace `yaml:"divorce,omitempty"`
	ChildrenCount     *uint     `yaml:"children_count,omitempty"`

	Person FlatPerson `yaml:"-"`
}

func (f FlatRelationship) GetPartner() (Person, error) {
	partner, err := f.Person.Database.GetByID(f.PartnerID)
	if err != nil {
		return nil, err
	}
	return partner, nil
}

func (f FlatRelationship) GetPartnerOrDummy() Person {
	partner, err := f.GetPartner()
	if err != nil {
		return NewDummyFlatPerson()
	}
	return partner
}

func (f FlatRelationship) GetIndex() int {
	return f.Index
}

func (f *FlatRelationship) SetChildrenCount(childrenCount *uint) {
	f.ChildrenCount = childrenCount
}

func (f FlatRelationship) GetChildrenCount() *uint {
	return f.ChildrenCount
}

func (f FlatRelationship) HasChildrenCount() bool {
	return f.ChildrenCount != nil
}

func (f FlatRelationship) GetEngagement() *DatePlace {
	return &f.Engagement
}

func (f FlatRelationship) GetMarriage() *DatePlace {
	return &f.Marriage
}

func (f FlatRelationship) GetMarriageReligious() *DatePlace {
	return &f.MarriageReligious
}

func (f FlatRelationship) GetDivorce() *DatePlace {
	return &f.Divorce
}

func (r FlatRelationship) GetEngagementDuration(a Person, referenceDate time.Time) timecalc.Duration {
	return r.getDurationInternal(r.GetEngagement(), r.GetMarriage(), 100, a, referenceDate)
}

func (r FlatRelationship) GetMarriageDuration(a Person, referenceDate time.Time) timecalc.Duration {
	// TODO handle divorce!
	return r.getDurationInternal(r.GetMarriage(), nil, 100, a, referenceDate)
}

func (r FlatRelationship) GetMarriageReligiousDuration(a Person, referenceDate time.Time) timecalc.Duration {
	return r.getDurationInternal(r.GetMarriageReligious(), nil, 100, a, referenceDate)
}

func (r FlatRelationship) getDurationInternal(start *DatePlace, end *DatePlace, maxYears int, a Person, referenceDate time.Time) timecalc.Duration {
	result := timecalc.Duration{}
	// get start
	if start.Date == "" || timecalc.SparseDateFrom(start.Date).IsEmpty() {
		result.Error = errors.Errorf("no marriage/no valid start")
		return result
	}
	s := *dateMax(start.Date)
	if s.IsZero() {
		result.Error = errors.Errorf("no marriage/no valid start")
		return result
	}
	// get end
	var e time.Time
	if end == nil {
		b, err := r.GetPartner()
		if err != nil {
			result.Error = err
			return result
		}
		if HasDied(a, referenceDate) {
			if a.GetDeath().Date == "" {
				result.Error = errors.Errorf("died, but no death date given")
				return result
			}
			min := dateMin(a.GetDeath().Date)
			e = *min
			// TODO result.IsMinimum = true
			result.IsFinished = true
		}
		if HasDied(b, referenceDate) {
			if b.GetDeath().Date == "" {
				result.Error = errors.Errorf("partner died, but no death date given")
				return result
			}
			deathB := dateMin(b.GetDeath().Date)
			if e.IsZero() || deathB.Before(e) {
				e = *deathB
			}
			// TODO result.IsMinimum = true
			result.IsFinished = true
		}
	} else {
		result.IsFinished = true
		e = *dateMin(end.Date)
	}
	if e.IsZero() {
		if referenceDate.IsZero() {
			result.Error = errors.Errorf("no end date")
			return result
		}
		e = referenceDate
	}

	diff := timecalc.FullDate{Time: &s}.Age(e)
	result.Years = diff.Years
	result.Months = diff.Months
	result.Days = diff.Days

	if diff.Years > maxYears {
		result.Error = errors.Errorf("duration exceeds maxYears of %d", maxYears)
		return result
	}

	return result
}
