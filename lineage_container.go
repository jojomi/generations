package generations

import (
	"fmt"
)

type LineageContainer struct {
	internal map[string]PersonList
}

func NewLineageContainer() *LineageContainer {
	return &LineageContainer{
		internal: make(map[string]PersonList, 0),
	}
}

func (l *LineageContainer) Get(ID string) *PersonList {
	if v, ok := l.internal[ID]; ok {
		return &v
	}
	return nil
}

// Add adds two entries. One with the key, one with the TreePositionID that is calculated on the fly
func (l *LineageContainer) Add(ID, treePositionID string, personList PersonList) *LineageContainer {
	//l.internal[ID] = personList
	l.internal[treePositionID] = personList
	return l
}

func (l *LineageContainer) AddWithBase(ID, treePositionID, baseID string, person Person) *LineageContainer {
	newLineage := NewPersonList()

	base := l.Get(baseID)
	if base != nil {
		newLineage.AddList(base)
	}

	newLineage.AddPerson(person)
	l.Add(ID, treePositionID, newLineage)
	return l
}

func (l *LineageContainer) String() string {
	result := ""
	for k, pL := range l.internal {
		result += k + ": "
		for _, p := range pL.GetPersons() {
			result += fmt.Sprintf("%s, ", p)
		}
		result += "\n"
	}
	return result
}
