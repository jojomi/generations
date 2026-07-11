package generations

import (
	"time"

	"github.com/jojomi/timecalc"
)

type Relationship interface {
	GetIndex() int
	GetPartner() (Person, error)
	GetPartnerOrDummy() Person
	GetEngagement() *DatePlace
	GetMarriage() *DatePlace
	GetMarriageReligious() *DatePlace
	GetDivorce() *DatePlace
	SetChildrenCount(childrenCount *uint)
	GetChildrenCount() *uint
	HasChildrenCount() bool

	GetEngagementDuration(Person, time.Time) timecalc.Duration
	GetMarriageDuration(Person, time.Time) timecalc.Duration
	GetMarriageReligiousDuration(Person, time.Time) timecalc.Duration
}
