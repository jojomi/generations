package generations

type Relationship interface {
	GetPartner() (Person, error)
	GetEngagement() *DatePlace
	GetMarriage() *DatePlace
	GetDivorce() *DatePlace
}
