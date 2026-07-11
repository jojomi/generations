package generations

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func boolPtr(b bool) *bool {
	return &b
}

func testDate(value string) time.Time {
	d, err := time.Parse("2006-01-02", value)
	if err != nil {
		panic(err)
	}
	return d
}

func TestIsBorn(t *testing.T) {
	ref := testDate("2020-06-15")

	// no birth date: assumed born
	assert.True(t, IsBorn(&FlatPerson{}, ref))

	// born in the past
	assert.True(t, IsBorn(&FlatPerson{Birth: DatePlace{Date: "1990-01-01"}}, ref))

	// born in the future
	assert.False(t, IsBorn(&FlatPerson{Birth: DatePlace{Date: "2021-01-01"}}, ref))
}

func TestHasDied(t *testing.T) {
	ref := testDate("2020-06-15")

	// explicit dead state wins
	assert.True(t, HasDied(&FlatPerson{Dead: boolPtr(true)}, ref))
	assert.False(t, HasDied(&FlatPerson{Dead: boolPtr(false), Death: DatePlace{Date: "1990-01-01"}}, ref))

	// no information: not dead
	assert.False(t, HasDied(&FlatPerson{}, ref))

	// death date in the past
	assert.True(t, HasDied(&FlatPerson{Death: DatePlace{Date: "1990-01-01"}}, ref))

	// death date after reference date
	assert.False(t, HasDied(&FlatPerson{Death: DatePlace{Date: "2021-01-01"}}, ref))
}

func TestIsAlive(t *testing.T) {
	ref := testDate("2020-06-15")

	// explicit dead state wins
	assert.False(t, IsAlive(&FlatPerson{Dead: boolPtr(true)}, ref))
	assert.True(t, IsAlive(&FlatPerson{Dead: boolPtr(false)}, ref))

	// recently born, no death: alive
	assert.True(t, IsAlive(&FlatPerson{Birth: DatePlace{Date: "1990-01-01"}}, ref))

	// died
	assert.False(t, IsAlive(&FlatPerson{
		Birth: DatePlace{Date: "1900-01-01"},
		Death: DatePlace{Date: "1980-01-01"},
	}, ref))

	// born more than 120 years ago: assumed dead
	assert.False(t, IsAlive(&FlatPerson{Birth: DatePlace{Date: "1850-01-01"}}, ref))

	// married more than 100 years ago: assumed dead
	assert.False(t, IsAlive(&FlatPerson{
		Partners: []FlatRelationship{
			{Marriage: DatePlace{Date: "1900-05-01"}},
		},
	}, ref))
}

func TestIsMarried(t *testing.T) {
	ref := testDate("2020-06-15")

	assert.False(t, IsMarried(&FlatPerson{}, ref))
	assert.False(t, IsMarried(&FlatPerson{
		Partners: []FlatRelationship{{}},
	}, ref))
	assert.True(t, IsMarried(&FlatPerson{
		Partners: []FlatRelationship{
			{Marriage: DatePlace{Date: "1990-05-01"}},
		},
	}, ref))
	assert.True(t, IsMarried(&FlatPerson{
		Partners: []FlatRelationship{
			{MarriageReligious: DatePlace{Date: "1990-05-01"}},
		},
	}, ref))
}

func TestGetLifeAge(t *testing.T) {
	ref := testDate("2020-06-15")

	// no birth date: no age
	_, err := GetLifeAge(&FlatPerson{}, ref)
	assert.NotNil(t, err)

	// not yet born
	_, err = GetLifeAge(&FlatPerson{Birth: DatePlace{Date: "2021-01-01"}}, ref)
	assert.NotNil(t, err)

	// alive: age at reference date
	age, err := GetLifeAge(&FlatPerson{Birth: DatePlace{Date: "1990-01-01"}}, ref)
	assert.Nil(t, err)
	assert.Equal(t, 30, age.Years)

	// dead: age at death
	age, err = GetLifeAge(&FlatPerson{
		Birth: DatePlace{Date: "1900-01-01"},
		Death: DatePlace{Date: "1980-01-01"},
	}, ref)
	assert.Nil(t, err)
	assert.Equal(t, 80, age.Years)

	// dead without any death information: no age
	_, err = GetLifeAge(&FlatPerson{
		Birth: DatePlace{Date: "1990-01-01"},
		Dead:  boolPtr(true),
	}, ref)
	assert.NotNil(t, err)

	// death known, but without date: no age
	_, err = GetLifeAge(&FlatPerson{
		Birth: DatePlace{Date: "1990-01-01"},
		Death: DatePlace{Location: Location{Place: "Musterstadt"}},
	}, ref)
	assert.NotNil(t, err)
}
