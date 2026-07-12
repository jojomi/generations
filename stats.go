package generations

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/jojomi/timecalc"
)

type Stats struct {
	referenceDate time.Time
}

func NewStats(referenceDate time.Time) *Stats {
	return &Stats{
		referenceDate: referenceDate,
	}
}

func (s *Stats) GetReferenceDate() time.Time {
	return s.referenceDate
}

func (s *Stats) MaleCount(personList *PersonList) int {
	filterFunc := func(p Person) bool {
		return p.GetGender() == GenderMale
	}
	return len(personList.GetFiltered(filterFunc).GetPersons())
}

func (s *Stats) FemaleCount(personList *PersonList) int {
	filterFunc := func(p Person) bool {
		return p.GetGender() == GenderFemale
	}
	return len(personList.GetFiltered(filterFunc).GetPersons())
}

func (s *Stats) AlivePersons(personList *PersonList) *PersonList {
	filterFunc := func(p Person) bool {
		return IsAlive(p, s.GetReferenceDate())
	}
	return personList.GetFiltered(filterFunc)
}

func (s *Stats) AliveCount(personList *PersonList) int {
	return len(s.AlivePersons(personList).GetPersons())
}

func (s *Stats) DeadCount(personList *PersonList) int {
	filterFunc := func(p Person) bool {
		return !IsAlive(p, s.GetReferenceDate())
	}
	return len(personList.GetFiltered(filterFunc).GetPersons())
}

func (s *Stats) MarriedCount(personList *PersonList) int {
	filterFunc := func(p Person) bool {
		return IsMarried(p, s.GetReferenceDate())
	}
	return len(personList.GetFiltered(filterFunc).GetPersons())
}

func (s *Stats) ChildrenCount(personList *PersonList) *RankingList {
	makeBirthRankingItem := func(p Person) []*RankingItem {
		count := p.GetChildrenCount()

		ri := RankingItem{}
		pList := PersonList{}
		pList.AddPerson(p)
		ri.Persons = &pList
		ri.Value = count
		ri.RankingValue = count

		return []*RankingItem{&ri}
	}
	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(int) > (b.RankingValue.(int))
	}
	return getRanking(personList, makeBirthRankingItem, comparatorFunc)
}

type RankingList struct {
	RankingItems []RankingItem
	Comparator   func(a, b RankingItem) bool
	mutex        *sync.Mutex
}

func (r *RankingList) AddItem(item RankingItem) *RankingList {
	r.RankingItems = append(r.RankingItems, item)
	return r // chaining
}

func (r *RankingList) GetItems() []RankingItem {
	return r.RankingItems
}

func (r *RankingList) GetTop(i int) []RankingItem {
	if len(r.RankingItems) <= i {
		return r.RankingItems
	}
	return r.RankingItems[:i]
}

func (r *RankingList) Len() int {
	return len(r.RankingItems)
}

func (r *RankingList) Less(i, j int) bool {
	return r.Comparator(r.RankingItems[i], r.RankingItems[j])
}

func (r *RankingList) Swap(i, j int) {
	r.RankingItems[i], r.RankingItems[j] = r.RankingItems[j], r.RankingItems[i]
}

func (r *RankingList) Lock() {
	r.mutex.Lock()
}

func (r *RankingList) Unlock() {
	r.mutex.Unlock()
}

func NewRankingList() *RankingList {
	rl := RankingList{
		RankingItems: make([]RankingItem, 0),
		mutex:        &sync.Mutex{},
	}
	return &rl
}

type RankingItem struct {
	Rank         int
	Value        interface{}
	RankingValue interface{}
	Persons      *PersonList
}

func (s *Stats) OldestBirths(personList *PersonList) *RankingList {
	makeBirthRankingItem := func(p Person) []*RankingItem {
		datePlace := p.GetBirth()
		if datePlace.Date == "" {
			return nil
		}
		dateSparse := timecalc.SparseDateFrom(datePlace.Date)
		if dateSparse.IsEmpty() {
			return nil
		}

		ri := RankingItem{}
		pList := PersonList{}
		pList.AddPerson(p)
		ri.Persons = &pList
		ri.Value = leadingDotsRegexp.ReplaceAllString(dateSparse.Format("02.01.2006"), "")
		ri.RankingValue = dateSparse.GetMaxDate()

		return []*RankingItem{&ri}
	}
	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(time.Time).Before(b.RankingValue.(time.Time))
	}
	return getRanking(personList, makeBirthRankingItem, comparatorFunc)
}

func (s *Stats) LongestMarriages(personList *PersonList) *RankingList {
	makeMarriageRankingItems := func(p Person) []*RankingItem {
		result := make([]*RankingItem, 0)
		referenceDate := s.GetReferenceDate()
		for _, r := range p.GetRelationships() {
			diff := r.GetMarriageDuration(p, referenceDate)
			diffRel := r.GetMarriageReligiousDuration(p, referenceDate)
			if diff.Error != nil && diffRel.Error != nil {
				continue
			}
			// if diff == nil {
			// 	diff = diffRel
			// } else {
			// 	if diffRel != nil {
			if diff.Years < diffRel.Years ||
				(diff.Years == diffRel.Years && diff.Months < diffRel.Months) ||
				(diff.Years == diffRel.Years && diff.Months == diffRel.Months && diff.Days < diffRel.Days) {
				diff = diffRel
			}
			// 	}
			// }

			ri := RankingItem{}
			pList := PersonList{}
			pList.AddPerson(p)
			partner, err := r.GetPartner()
			if err != nil {
				panic(err)
			}
			pList.AddPerson(partner)
			ri.Persons = &pList

			ri.RankingValue = float64(diff.Years*365 + 30*diff.Months + diff.Days)
			ri.Value = formatDurationGerman(diff)

			result = append(result, &ri)
		}

		if len(result) == 0 {
			return nil
		}
		return result
	}
	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(float64) > b.RankingValue.(float64)
	}

	ranked := getRanking(personList, makeMarriageRankingItems, comparatorFunc)
	result := RankingList{}
outer:
	for _, r := range ranked.RankingItems {
		// always add the female first variant
		if r.Persons.persons[0].GetGender() != GenderMale {
			result.AddItem(r)
			continue
		}

		for _, r2 := range ranked.RankingItems {
			if r.Persons.persons[0].GetBestID() == r2.Persons.persons[1].GetBestID() && r.Persons.persons[1].GetBestID() == r2.Persons.persons[0].GetBestID() {
				continue outer
			}
		}
		result.AddItem(r)
	}

	return &result
}

func (s *Stats) NewestBirths(personList *PersonList) *RankingList {
	makeBirthRankingItem := func(p Person) []*RankingItem {
		datePlace := p.GetBirth()
		if datePlace.Date == "" {
			return nil
		}
		dateSparse := timecalc.SparseDateFrom(datePlace.Date)
		if dateSparse.IsEmpty() {
			return nil
		}

		ri := RankingItem{}
		pList := PersonList{}
		pList.AddPerson(p)
		ri.Persons = &pList
		ri.Value = leadingDotsRegexp.ReplaceAllString(dateSparse.Format("02.01.2006"), "")
		ri.RankingValue = dateSparse.GetMinDate()

		return []*RankingItem{&ri}
	}

	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(time.Time).After(b.RankingValue.(time.Time))
	}
	return getRanking(personList, makeBirthRankingItem, comparatorFunc)
}

func (s *Stats) Oldest(personList *PersonList) *RankingList {
	referenceTime := s.GetReferenceDate()
	makeAgeRankingItem := func(p Person) []*RankingItem {
		diff, err := GetLifeAge(p, referenceTime)
		if err != nil {
			return nil
		}
		dead := !IsAlive(p, referenceTime)

		pList := PersonList{}
		pList.AddPerson(p)
		ri := RankingItem{}
		ri.Persons = &pList
		ri.RankingValue = float64(diff.Years*365 + 30*diff.Months + diff.Days)
		ri.Value = fmt.Sprintf("%d Jahre, %d Monate, %d Tage", diff.Years, diff.Months, diff.Days)
		if dead {
			ri.Value = ri.Value.(string) + " (verstorben)"
		}
		return []*RankingItem{&ri}
	}

	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(float64) > b.RankingValue.(float64)
	}
	return getRanking(personList, makeAgeRankingItem, comparatorFunc)
}

func (s *Stats) LongestName(personList *PersonList) *RankingList {
	makeAgeRankingItem := func(p Person) []*RankingItem {
		name := p.GetName()

		//birthName := name.GetUsedFirst() + " " + name.Birth
		currentName := name.GetUsedFirst() + " " + name.Last
		shortName := currentName

		/*shortName := birthName
		if getNameLength(currentName) >= getNameLength(shortName) {
			shortName = currentName
		}*/

		pList := PersonList{}
		pList.AddPerson(p)
		nameLength := getNameLength(shortName)
		ri := RankingItem{}
		ri.Persons = &pList
		ri.RankingValue = nameLength
		ri.Value = fmt.Sprintf("%s (%d Buchstaben)", shortName, nameLength)
		return []*RankingItem{&ri}
	}

	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(int) > b.RankingValue.(int)
	}
	return getRanking(personList, makeAgeRankingItem, comparatorFunc)
}

func (s *Stats) TotalAgeLiving(personList *PersonList) int64 {
	var (
		sum           int64
		referenceDate = s.GetReferenceDate()
	)
	for _, p := range personList.GetPersons() {
		if !IsAlive(p, referenceDate) {
			continue
		}
		age, err := GetLifeAge(p, referenceDate)
		if err != nil {
			// age unknown, cannot contribute to the sum
			continue
		}
		sum += int64(age.Years)
	}
	return sum
}

func (s *Stats) NonUniqueNames(personList *PersonList) *RankingList {
	getKey := func(p Person) string {
		name := p.GetName()
		if getNameLength(name.GetUsedFirst()) < 2 {
			return ""
		}
		if getNameLength(name.Last) < 2 {
			return ""
		}
		fullName := name.FormatFullNoMiddleNoBirth()
		if fullName == "" {
			return ""
		}
		return fullName
	}
	histogram := getHistogram(personList, getKey)

	// filter
	hist2 := getHistogramNonUnique(filterHistogramByBestID(histogram))

	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(int) > b.RankingValue.(int)
	}
	makeRankingItem := func(k string, v *PersonList) RankingItem {
		return RankingItem{
			Persons:      v,
			Value:        k,
			RankingValue: v.Count(),
		}
	}
	return getRankingForHistogram(hist2, comparatorFunc, makeRankingItem)
}

func (s *Stats) ShortestName(personList *PersonList) *RankingList {
	makeAgeRankingItem := func(p Person) []*RankingItem {
		name := p.GetName()

		usedFirst := name.GetUsedFirst()
		//birthName := usedFirst + " " + name.Birth
		currentName := usedFirst + " " + name.Last

		if getNameLength(usedFirst) < 2 || getNameLength(name.Last) < 2 {
			return nil
		}
		shortName := currentName

		/*if getNameLength(usedFirst) > 1 && getNameLength(name.Birth) > 1 {
			shortName = birthName
		}
		if getNameLength(usedFirst) > 1 && getNameLength(name.Last) > 1 && getNameLength(currentName) <= getNameLength(shortName) {
			shortName = currentName
		}*/

		pList := PersonList{}
		pList.AddPerson(p)
		nameLength := getNameLength(shortName)
		ri := RankingItem{}
		ri.Persons = &pList
		ri.RankingValue = nameLength
		ri.Value = fmt.Sprintf("%s (%d Buchstaben)", shortName, nameLength)
		return []*RankingItem{&ri}
	}

	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(int) < b.RankingValue.(int)
	}
	return getRanking(personList, makeAgeRankingItem, comparatorFunc)
}

var nameReplacer = strings.NewReplacer(" ", "", "-", "", ".", "", "?", "")

var leadingDotsRegexp = regexp.MustCompile(`^\.+`)

func pluralSuffixE(count int) string {
	if count == 1 {
		return ""
	}
	return "e"
}

func formatDurationGerman(diff timecalc.Duration) string {
	return fmt.Sprintf("%d Jahr%s, %d Monat%s, %d Tag%s",
		diff.Years, pluralSuffixE(diff.Years),
		diff.Months, pluralSuffixE(diff.Months),
		diff.Days, pluralSuffixE(diff.Days))
}

func getNameLength(input string) int {
	filteredName := nameReplacer.Replace(input)
	return utf8.RuneCountInString(filteredName)
}

func getRanking(personList *PersonList, makeRankingItems func(p Person) []*RankingItem, comparatorFunc func(a, b RankingItem) bool) *RankingList {
	result := NewRankingList()
	result.Comparator = comparatorFunc
	var (
		rankingItems []*RankingItem
	)
	for _, p := range personList.GetPersons() {
		rankingItems = makeRankingItems(p)
	rankingItem:
		for _, r := range rankingItems {
			if r == nil {
				continue
			}
			// check if multiple instances of the same person need to be merged
			for _, res := range result.RankingItems {
				if r.Value == res.Value && r.Persons.HasSameElements(res.Persons) {
					res.Persons.MergeTreePositions(r.Persons)
					continue rankingItem
				}
			}

			result.AddItem(*r)
		}
	}
	sort.Sort(result)
	return result
}

func (s *Stats) MostCommonFirstNames(personList *PersonList) *RankingList {
	getKey := func(p Person) string {
		firstname := p.GetName().GetUsedFirst()
		if firstname == "" {
			return ""
		}
		return firstname
	}
	personList = personList.GetWithMergedTreePositionIDs()
	histogram := getHistogram(personList, getKey)
	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(int) > b.RankingValue.(int)
	}
	return getRankingForHistogram(histogram, comparatorFunc, makeRankingItemDefault)
}

func (s *Stats) MostCommonLastNames(personList *PersonList) *RankingList {
	getKey := func(p Person) string {
		lastname := p.GetName().Last
		if lastname == "" {
			return ""
		}
		return lastname
	}
	personList = personList.GetWithMergedTreePositionIDs()
	histogram := getHistogram(personList, getKey)
	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(int) > b.RankingValue.(int)
	}
	return getRankingForHistogram(histogram, comparatorFunc, makeRankingItemDefault)
}

func (s *Stats) UniqueFirstNames(personList *PersonList) *RankingList {
	getKey := func(p Person) string {
		firstname := p.GetName().GetUsedFirst()
		if firstname == "" {
			return ""
		}
		if getNameLength(firstname) < 2 {
			return ""
		}
		if strings.Contains(firstname, "/") {
			return ""
		}
		if strings.Contains(firstname, ".") {
			return ""
		}
		return firstname
	}
	histogram := getHistogram(personList, getKey)

	// filter
	hist2 := getHistogramUnique(histogram)

	comparatorFunc := func(a, b RankingItem) bool {
		return a.Value.(string) < b.Value.(string)
	}
	return getRankingForHistogram(hist2, comparatorFunc, makeRankingItemDefault)
}

func (s *Stats) DeathAges(personList *PersonList) *RankingList {
	referenceDate := s.GetReferenceDate()
	getKey := func(p Person) string {
		birth := p.GetBirth()
		if birth.Date == "" {
			return ""
		}
		if IsAlive(p, referenceDate) {
			return ""
		}
		age, err := GetLifeAge(p, referenceDate)
		if err != nil {
			return ""
		}
		return strconv.Itoa(age.Years)
	}
	personList = personList.GetWithMergedTreePositionIDs()
	histogram := getHistogram(personList, getKey)

	comparatorFunc := func(a, b RankingItem) bool {
		aInt, _ := strconv.Atoi(a.Value.(string))
		bInt, _ := strconv.Atoi(b.Value.(string))
		return aInt < bInt
	}
	makeRankingItem := func(k string, v *PersonList) RankingItem {
		date, _ := time.Parse("01", k)
		return RankingItem{
			Persons:      v,
			Value:        k,
			RankingValue: date.Format("01"),
		}
	}
	return getRankingForHistogram(histogram, comparatorFunc, makeRankingItem)
}

func getMarriageAgeHistogram(personList *PersonList) map[string]*PersonList {
	getKey := func(p Person) string {
		relationships := p.GetRelationships()

		birth := p.GetBirth()
		if birth.Date == "" {
			return ""
		}
		dateBirth := timecalc.SparseDateFrom(birth.Date)
		if dateBirth == nil || !dateBirth.IsComplete() {
			return ""
		}
		minBirth := dateBirth.GetMinDate()

		// nur erste Hochzeit beachten
		if len(relationships) == 0 {
			return ""
		}
		r := relationships[0]
		var m *DatePlace
		m = r.GetMarriage()
		if m.Date == "" || timecalc.SparseDateFrom(m.Date) == nil {
			m = r.GetMarriageReligious()
		}
		if m.Date == "" {
			return ""
		}
		date := timecalc.SparseDateFrom(m.Date)
		if date == nil || !date.IsComplete() {
			return ""
		}
		latestMarriageDate := date.GetMaxDate()
		diff := timecalc.FullDate{Time: &minBirth}.Age(latestMarriageDate)
		if diff.Years > 110 {
			return ""
		}
		return strconv.Itoa(diff.Years)
	}
	histogram := getHistogram(personList, getKey)
	return histogram
}

func (s *Stats) AvgMarriageAge(personList *PersonList) float64 {
	histogram := getMarriageAgeHistogram(personList)

	sum := 0.0
	count := 0
	for key, personList := range histogram {
		age, err := strconv.Atoi(key)
		if err != nil {
			panic(err)
		}
		count += personList.Count()
		sum += float64(age * personList.Count())
	}

	if count == 0 {
		return 0
	}
	return sum / float64(count)
}

func (s *Stats) MarriageAges(personList *PersonList) *RankingList {
	personList = personList.GetWithMergedTreePositionIDs()
	histogram := getMarriageAgeHistogram(personList)

	comparatorFunc := func(a, b RankingItem) bool {
		aInt, _ := strconv.Atoi(a.Value.(string))
		bInt, _ := strconv.Atoi(b.Value.(string))
		return aInt < bInt
	}
	makeRankingItem := func(k string, v *PersonList) RankingItem {
		date, _ := time.Parse("01", k)
		return RankingItem{
			Persons:      v,
			Value:        k,
			RankingValue: date.Format("01"),
		}
	}
	return getRankingForHistogram(histogram, comparatorFunc, makeRankingItem)
}

func (s *Stats) BirthDays(personList *PersonList) *RankingList {
	sortFormat := "0102"
	getKey := func(p Person) string {
		birth := p.GetBirth()
		if birth.Date == "" {
			return ""
		}
		birthSparse := timecalc.SparseDateFrom(birth.Date)
		if !birthSparse.IsCompleteDayAndMonth() {
			return ""
		}
		monthDay := birthSparse.Format(sortFormat)
		return monthDay
	}
	personList = personList.GetWithMergedTreePositionIDs()
	histogram := getHistogram(personList, getKey)

	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(string) < b.RankingValue.(string)
	}
	makeRankingItem := func(k string, v *PersonList) RankingItem {
		date, _ := time.Parse(sortFormat, k)
		return RankingItem{
			Persons:      v,
			Value:        date.Format("02.01."),
			RankingValue: k,
		}
	}
	return getRankingForHistogram(histogram, comparatorFunc, makeRankingItem)
}

func (s *Stats) DeathDays(personList *PersonList) *RankingList {
	sortFormat := "0102"
	getKey := func(p Person) string {
		birth := p.GetDeath()
		if birth.Date == "" {
			return ""
		}
		birthSparse := timecalc.SparseDateFrom(birth.Date)
		if !birthSparse.IsCompleteDayAndMonth() {
			return ""
		}
		monthDay := birthSparse.Format(sortFormat)
		return monthDay
	}
	personList = personList.GetWithMergedTreePositionIDs()
	histogram := getHistogram(personList, getKey)

	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(string) < b.RankingValue.(string)
	}
	makeRankingItem := func(k string, v *PersonList) RankingItem {
		date, _ := time.Parse(sortFormat, k)
		return RankingItem{
			Persons:      v,
			Value:        date.Format("02.01."),
			RankingValue: k,
		}
	}
	return getRankingForHistogram(histogram, comparatorFunc, makeRankingItem)
}

func (s *Stats) MarriageDays(personList *PersonList) *RankingList {
	sortFormat := "01022006"
	makeMarriageRankingItems := func(p Person) []*RankingItem {
		result := make([]*RankingItem, 0)
		for _, r := range p.GetRelationships() {
			marriage := r.GetMarriage()
			if !timecalc.SparseDateFrom(marriage.Date).IsCompleteDayAndMonth() {
				marriage = r.GetMarriageReligious()
			}

			if marriage.Date == "" {
				continue
			}
			marriageSparse := timecalc.SparseDateFrom(marriage.Date)
			if !marriageSparse.IsCompleteDayAndMonth() {
				continue
			}

			ri := RankingItem{}
			pList := PersonList{}
			pList.AddPerson(p)
			partner := r.GetPartnerOrDummy()
			pList.AddPerson(partner)
			ri.Persons = &pList

			ri.RankingValue = marriageSparse.Format(sortFormat)
			ri.Value = marriageSparse

			result = append(result, &ri)
		}

		if len(result) == 0 {
			return nil
		}
		return result
	}
	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(string) < b.RankingValue.(string)
	}

	ranked := getRanking(personList, makeMarriageRankingItems, comparatorFunc)
	result := RankingList{}
outer:
	for _, r := range ranked.RankingItems {
		// always add the female first variant
		if r.Persons.persons[0].GetGender() != GenderMale {
			result.AddItem(r)
			continue
		}

		for _, r2 := range ranked.RankingItems {
			if r.Persons.persons[0].GetBestID() == r2.Persons.persons[1].GetBestID() && r.Persons.persons[1].GetBestID() == r2.Persons.persons[0].GetBestID() {
				continue outer
			}
		}
		result.AddItem(r)
	}

	return &result
}

func (s *Stats) BirthMonths(personList *PersonList) *RankingList {
	getKey := func(p Person) string {
		birth := p.GetBirth()
		if birth.Date == "" {
			return ""
		}
		month := timecalc.SparseDateFrom(birth.Date).Format("01")
		if strings.Contains(month, "?") {
			return ""
		}
		return month
	}
	personList = personList.GetWithMergedTreePositionIDs()
	histogram := getHistogram(personList, getKey)

	comparatorFunc := func(a, b RankingItem) bool {
		return a.Value.(string) < b.Value.(string)
	}
	makeRankingItem := func(k string, v *PersonList) RankingItem {
		date, _ := time.Parse("01", k)
		return RankingItem{
			Persons:      v,
			Value:        k,
			RankingValue: date.Format("01"),
		}
	}
	return getRankingForHistogram(histogram, comparatorFunc, makeRankingItem)
}

func (s *Stats) NonUniqueBirthdays(personList *PersonList) *RankingList {
	getKey := func(p Person) string {
		birth := p.GetBirth()
		if birth.Date == "" {
			return ""
		}
		birthDate := timecalc.SparseDateFrom(birth.Date)
		if !birthDate.IsCompleteDayAndMonth() {
			return ""
		}
		return birthDate.Format("01.02")
	}
	personList = personList.GetWithMergedTreePositionIDs()
	histogram := getHistogram(personList, getKey)

	// filter
	hist2 := getHistogramNonUnique(histogram)

	comparatorFunc := func(a, b RankingItem) bool {
		return a.Value.(string) < b.Value.(string)
	}
	makeRankingItem := func(k string, v *PersonList) RankingItem {
		date, _ := time.Parse("01.02", k)
		return RankingItem{
			Persons:      v,
			Value:        k,
			RankingValue: date.Format("02.01."),
		}
	}
	return getRankingForHistogram(hist2, comparatorFunc, makeRankingItem)
}

func (s *Stats) MultipleAncestors(personList *PersonList) *RankingList {
	makeRankingItems := func(p Person) []*RankingItem {
		ids := p.GetTreePositionIDs()
		if len(ids) < 2 {
			return nil
		}
		// check if at least one position is an ancestor (depth >= 2)
		isAncestor := false
		for _, id := range ids {
			if strings.Count(id, "-") >= 2 {
				isAncestor = true
				break
			}
		}
		if !isAncestor {
			return nil
		}

		ri := RankingItem{}
		pList := PersonList{}
		pList.AddPerson(p)
		ri.Persons = &pList
		ri.Value = len(ids)
		ri.RankingValue = len(ids)

		return []*RankingItem{&ri}
	}
	comparatorFunc := func(a, b RankingItem) bool {
		return a.RankingValue.(int) > b.RankingValue.(int)
	}
	return getRanking(personList, makeRankingItems, comparatorFunc)
}

func getHistogramUnique(histogram map[string]*PersonList) map[string]*PersonList {
	result := make(map[string]*PersonList, len(histogram)/2)
	for key, value := range histogram {
		if value.Count() != 1 {
			continue
		}
		result[key] = value
	}
	return result
}

func getHistogramNonUnique(histogram map[string]*PersonList) map[string]*PersonList {
	result := make(map[string]*PersonList, len(histogram)/2)
	for key, value := range histogram {
		if value.Count() < 2 {
			continue
		}
		result[key] = value
	}
	return result
}

func filterHistogramByBestID(input map[string]*PersonList) map[string]*PersonList {
	for k, v := range input {
		input[k] = v.RemoveDuplicatesByBestID()
	}
	return input
}

func getHistogram(personList *PersonList, getItemKey func(p Person) string) map[string]*PersonList {
	result := make(map[string]*PersonList, 0)
	var key string
	for _, p := range personList.GetPersons() {
		key = getItemKey(p)
		if key == "" {
			continue
		}
		if _, ok := result[key]; !ok {
			result[key] = &PersonList{}
		}
		result[key].AddPerson(p)
	}
	return result
}

func getRankingForHistogram(histogram map[string]*PersonList, comparatorFunc func(a, b RankingItem) bool, makeRankingItem func(k string, v *PersonList) RankingItem) *RankingList {
	result := NewRankingList()
	for k, v := range histogram {
		r := makeRankingItem(k, v)
		result.AddItem(r)
	}
	result.Lock()
	result.Comparator = comparatorFunc
	sort.Sort(result)
	result.Unlock()
	return result
}

func makeRankingItemDefault(k string, v *PersonList) RankingItem {
	return RankingItem{
		Persons:      v,
		Value:        k,
		RankingValue: v.Count(),
	}
}
