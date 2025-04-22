package report

import (
	"fmt"
	"io"

	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
)

const (
	blankName     = " "
	spacer        = " "
	fieldSizeProj = 5
)

func Cal(
	w io.Writer,
	project string,
	epicMap map[myj.MyKey]*myj.ResponseIssue,
	outer *utils.DayRange,
	color bool,
	fieldSizeName int,
	showHeaders bool,
	lineSetSize int,
) {
	epicKeys := myj.GetSortedKeys(epicMap)
	fmProj := fmt.Sprintf("%%%ds", fieldSizeProj)
	fmId := fmt.Sprintf("%%%dd", fieldSizeProj)
	fmName := fmt.Sprintf("%%%ds", fieldSizeName)
	if showHeaders {
		_, _ = fmt.Fprintf(w, fmProj, blankName)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintf(w, fmName, blankName)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintln(w, outer.MonthHeader())

		h1, h2 := outer.DayHeaders()
		_, _ = fmt.Fprintf(w, fmProj, blankName)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintf(w, fmName, blankName)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintln(w, h1)

		_, _ = fmt.Fprintf(w, fmProj, project)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintf(w, fmName, blankName)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintln(w, h2)
	}
	today := utils.Today()
	lineCount := 0
	for _, epicKey := range epicKeys {
		epic := epicMap[epicKey.MyKey]
		dr, err := utils.MakeDayRange0(epic.DateStart(), epic.DateEnd())
		if err != nil {
			panic(err)
		}
		_, _ = fmt.Fprintf(w, fmId, epic.MyKey.Num)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintf(w, fmName, utils.Ellipsis(epic.MySummary(), fieldSizeName))
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprint(w, dr.AsIntersect(today, outer, color))
		_, _ = fmt.Fprintln(w)
		lineCount++
		if lineCount%lineSetSize == 0 {
			_, _ = fmt.Fprintln(w)
		}
	}
}
