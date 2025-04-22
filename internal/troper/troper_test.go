package troper

import (
	"bufio"
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/report"
	"github.com/monopole/gojira/internal/utils"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

const (
	RW os.FileMode = 0644
)

func TestUnSpewEpics(t *testing.T) {
	expectedIntermediates := []*ParsedJiraLine{
		{
			IsEpic: true,
			ParsedIssue: ParsedIssue{
				Proj:      "BUDS",
				Num:       598,
				Type:      myj.IssueTypeEpic,
				Status:    myj.IssueStatusBacklog,
				Start:     utils.MakeDate(2025, time.March, 3),
				End:       utils.MakeDate(2025, time.April, 14),
				RawLabels: []string{"blah"},
				Summary:   "Sirius sirius",
			},
		},
		{
			ParsedIssue: ParsedIssue{
				Proj:      "BUDS",
				Num:       607,
				Type:      myj.IssueTypeTask,
				Status:    myj.IssueStatusInProgress,
				Start:     utils.MakeDate(2025, time.March, 6),
				End:       utils.MakeDate(2025, time.March, 19),
				RawLabels: []string{"blah"},
				Summary:   "Procyon procyon",
			},
		},
		{
			ParsedIssue: ParsedIssue{
				Proj:    "CIA",
				Num:     606,
				Type:    myj.IssueTypeStory,
				Status:  myj.IssueStatusDone,
				Start:   utils.MakeDate(2025, time.March, 20),
				End:     utils.MakeDate(2025, time.April, 17),
				Summary: "Rigel rigel",
			},
		},
		{
			ParsedIssue: ParsedIssue{
				Proj:      "BUDS",
				Num:       608,
				Type:      myj.IssueTypeTask,
				Status:    myj.IssueStatusClosedWoAction,
				Start:     utils.MakeDate(2025, time.April, 15),
				End:       utils.MakeDate(2025, time.June, 10),
				RawLabels: []string{"blah"},
				Summary:   "Arcturus arcturus",
			},
		},
		{
			IsEpic: true,
			ParsedIssue: ParsedIssue{
				Proj:    "BUDS",
				Num:     597,
				Type:    myj.IssueTypeEpic,
				Status:  myj.IssueStatusInQueue,
				Start:   utils.MakeDate(2026, time.February, 5),
				End:     utils.MakeDate(2026, time.February, 26),
				Summary: "Vega vega",
			},
		},
		{
			IsEpic: true,
			ParsedIssue: ParsedIssue{
				Proj:      "BUDS",
				Num:       596,
				Type:      myj.IssueTypeEpic,
				Status:    myj.IssueStatusInQueue,
				Start:     utils.MakeDate(2025, time.April, 16),
				End:       utils.MakeDate(2025, time.May, 21),
				RawLabels: []string{"blah"},
				Summary:   "Capella capella",
			},
		},
		{
			ParsedIssue: ParsedIssue{
				Proj:      "BUDS",
				Num:       605,
				Type:      myj.IssueTypeStory,
				Status:    myj.IssueStatusReadyForDev,
				Start:     utils.MakeDate(2025, time.June, 15),
				End:       utils.MakeDate(2025, time.August, 10),
				RawLabels: []string{"blah"},
				Summary:   "Rigel rigel",
			},
		},
		{
			ParsedIssue: ParsedIssue{
				Proj:      "BUDS",
				Num:       604,
				Type:      myj.IssueTypeTask,
				Status:    myj.IssueStatusBacklog,
				Start:     utils.MakeDate(2025, time.October, 15),
				End:       utils.MakeDate(2025, time.December, 10),
				RawLabels: []string{"ap8_-ple", "peach"},
				Summary:   "Betelgeuse betelgeuse",
			},
		},
	}
	const (
		inputData = `

BUDS-598 [Epic] (Backlog)  2025-Mar-03 2025-Apr-14  6w <blah> Sirius sirius

     BUDS-607 [Task] (In Progress) 2025-Mar-06 2025-Mar-19 2w <blah>  Procyon procyon
     CIA-606 [Story] (Done)   2025-Mar-20 2025-Apr-17 4w <> Rigel rigel
     BUDS-608 [Task] (Closed Without Action)   2025-Apr-15 2025-Jun-10 8w  <blah> Arcturus arcturus

Epic BUDS-597 [Epic] (In Queue)  2026-Feb-05 2026-Feb-26 3w <>           Vega vega

Epic BUDS-596 [Epic] (In Queue)  2025-Apr-16 2025-May-21 5w <blah> Capella capella

     BUDS-605 [Story] (Ready For Development)     2025-Jun-15 2025-Aug-10 8w    <blah>  Rigel rigel
     BUDS-604 [Task] (Backlog)  2025-Oct-15 2025-Dec-10 8w <ap8_-ple,peach>  Betelgeuse betelgeuse



`
		expectedFinal = `
BUDS-598      [Epic]       (Backlog)        2025-Mar-03 2025-Apr-14   6w <blah> Sirius sirius
  BUDS-607    [Task]       (In Progress)    2025-Mar-06 2025-Mar-19   2w <blah> Procyon procyon
  BUDS-608    [Task]       (Closed Withou…) 2025-Apr-15 2025-Jun-10   8w <blah> Arcturus arcturus
  CIA-606     [Story]      (Done)           2025-Mar-20 2025-Apr-17   4w <> Rigel rigel

BUDS-596      [Epic]       (In Queue)       2025-Apr-16 2025-May-21   5w <blah> Capella capella
  BUDS-604    [Task]       (Backlog)        2025-Oct-15 2025-Dec-10   8w <ap8_-ple,peach> Betelgeuse betelgeuse
  BUDS-605    [Story]      (Ready for Dev…) 2025-Jun-15 2025-Aug-10   8w <blah> Rigel rigel

BUDS-597      [Epic]       (In Queue)       2026-Feb-05 2026-Feb-26   3w <> Vega vega
`
	)

	// Create a file with slightly damaged but still parsable formatting.
	fs := afero.NewMemMapFs() // for real thing use: NewOsFs()
	const fName = `issues.txt`
	assert.NoError(t, afero.WriteFile(fs, fName, []byte(inputData), RW))

	// Read the file, compare to expected intermediate structures.
	actual, err := UnSpewEpics(fs, fName)
	assert.NoError(t, err)
	assert.Equal(t, expectedIntermediates, actual)

	// FromJiraOrDie the intermediate structures back to "native" types,
	// and write them using the Print function.
	// Confirm the formatting.
	em, im := Convert(actual)
	var b bytes.Buffer
	{
		w := bufio.NewWriter(&b)
		report.SpewEpics(w, em, im, getEpicLink)
		assert.NoError(t, w.Flush())
	}
	assert.Equal(t, expectedFinal[1:], b.String())
}

type fakeJb struct {
}

func (jb *fakeJb) Project() string {
	return "blah"
}

func (jb *fakeJb) GetOneIssue(_ int) (*myj.ResponseIssue, error) {
	return nil, nil
}

func getEpicLink(ir *myj.ResponseIssue) (result myj.MyKey) {
	str, ok := ir.Fields.CustomEpicLink.(string)
	if ok && str != "" {
		return myj.ParseMyKey(str)
	}
	return myj.MyKey{
		Proj: "UNKNOWN",
		Num:  9000,
	}
}
