package myj

import (
	"log"
	"sort"
	"strconv"
	"strings"

	"github.com/monopole/gojira/internal/utils"
)

type MyKey struct {
	Proj string `json:"proj,omitempty"`
	Num  int    `json:"num,omitempty"`
}

func (mk MyKey) String() string {
	return mk.Proj + "-" + strconv.Itoa(mk.Num)
}

func ParseMyKey(k string) (result MyKey) {
	parts := strings.Split(k, "-")
	if len(parts) != 2 {
		log.Fatalf("expected something like PEACH-1234, but see %q", k)
	}
	result.Proj = strings.ToUpper(parts[0])
	var err error
	result.Num, err = strconv.Atoi(parts[1])
	if err != nil {
		log.Fatalf(
			"Not number; expected something like PEACH-1234, but got %q", k)
	}
	return
}

type SrtKey struct {
	MyKey
	utils.Date
}

type KeyList []SrtKey

func (kl KeyList) Len() int {
	return len(kl)
}

func (kl KeyList) Less(i, j int) bool {
	return kl[i].Date.Before(kl[j].Date)
}

func (kl KeyList) Less2(i, j int) bool {
	if kl[i].Proj == kl[j].Proj {
		return kl[i].Num < kl[j].Num
	}
	return kl[i].Proj < kl[j].Proj
}

func (kl KeyList) Swap(i, j int) {
	kl[i], kl[j] = kl[j], kl[i]
}

func GetSortedKeys(m map[MyKey]*ResponseIssue) KeyList {
	keys := make(KeyList, len(m))
	{
		i := -1
		for k, v := range m {
			i++
			date := utils.Today()
			// Use start date
			d, err := utils.ParseDate(v.Fields.CustomStartDate)
			if err == nil {
				date = d
			}
			keys[i] = SrtKey{
				MyKey: k,
				Date:  date,
			}
		}
	}
	sort.Sort(keys)
	return keys
}
