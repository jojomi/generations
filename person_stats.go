package generations

import (
	"time"

	"github.com/jojomi/timecalc"
	"github.com/juju/errors"
)

func dateMin(raw string) *time.Time {
	minDate := timecalc.SparseDateFrom(raw).GetMinDate()
	return &minDate
}

func dateMax(raw string) *time.Time {
	maxDate := timecalc.SparseDateFrom(raw).GetMaxDate()
	return &maxDate
}

func IsBorn(p Person, referenceDate time.Time) bool {
	birth := p.GetBirth()
	if birth.Date == "" {
		return true
	}
	birthMin := dateMin(birth.Date)
	return !birthMin.After(referenceDate)
}

func HasDied(p Person, referenceDate time.Time) bool {
	// if set explicitly, use it
	if state := p.GetDeadState(); state != nil {
		return *state
	}

	// must not be dead
	death := p.GetDeath()
	if death.Date == "" {
		return false
	}
	maxDeath := dateMax(death.Date)
	if maxDeath.Before(referenceDate) {
		return true
	}
	return false
}

func IsAlive(p Person, referenceDate time.Time) bool {
	// if set explicitly, use it
	if state := p.GetDeadState(); state != nil {
		return !*state
	}

	if !IsBorn(p, referenceDate) || HasDied(p, referenceDate) {
		return false
	}
	// the contact is guaranteed to be born here, but there is no date of death
	maxBirth := dateMax(p.GetBirth().Date)
	if maxBirth.AddDate(120, 0, 0).Before(referenceDate) {
		return false
	}
	// more attempts to find out if the person is alive: Can't be alive if he/she married at least 100 years ago.
	for _, r := range p.GetRelationships() {
		maxDate := dateMax(r.GetMarriage().Date)
		if maxDate.AddDate(100, 0, 0).Before(referenceDate) {
			return false
		}
	}
	return true
}

func IsMarried(p Person, referenceDate time.Time) bool {
	for _, r := range p.GetRelationships() {
		if !r.GetMarriage().Empty() {
			return true
		}
		if !r.GetMarriageReligious().Empty() {
			return true
		}
	}
	return false
}

// GetLifeAge returns the minimal age
func GetLifeAge(p Person, referenceDate time.Time) (*timecalc.DateDiff, error) {
	if p.GetBirth().Date == "" {
		return nil, errors.Errorf("no birth date")
	}
	if !IsBorn(p, referenceDate) {
		// not yet born
		return nil, errors.Errorf("not born")
	}
	latestBirth := dateMax(p.GetBirth().Date)
	if latestBirth.Year() < 100 {
		return nil, errors.Errorf("birth unclear")
	}

	death := p.GetDeath()
	if !death.Empty() && death.Date == "" {
		return nil, errors.Errorf("dead, but no idea when")
	}
	if death.Empty() && !IsAlive(p, referenceDate) {
		return nil, errors.Errorf("dead, but no death given")
	}
	if death.Empty() {
		// born, but not died (possibly alive)
		diff := timecalc.FullDate{Time: latestBirth}.Age(referenceDate)
		return &diff, nil
	}
	// born and died
	earliestDeath := dateMin(death.Date)
	diff := timecalc.FullDate{Time: latestBirth}.Age(*earliestDeath)
	return &diff, nil
}
