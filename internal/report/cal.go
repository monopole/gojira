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

type CalParams struct {
	ProjectName   string
	Outer         *utils.DayRange
	UseColor      bool
	FieldSizeName int
	ShowHeaders   bool
	LineSetSize   int
}

func DoCal(
	w io.Writer,
	epicMap map[myj.MyKey]*myj.ResponseIssue,
	p CalParams,
) {
	epicKeys := myj.GetSortedKeys(epicMap)
	fmProj := fmt.Sprintf("%%%ds", fieldSizeProj)
	fmId := fmt.Sprintf("%%%dd", fieldSizeProj)
	fmName := fmt.Sprintf("%%%ds", p.FieldSizeName)
	if p.ShowHeaders {
		_, _ = fmt.Fprintf(w, fmProj, blankName)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintf(w, fmName, blankName)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintln(w, p.Outer.MonthHeader())

		h1, h2 := p.Outer.DayHeaders()
		_, _ = fmt.Fprintf(w, fmProj, blankName)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintf(w, fmName, blankName)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintln(w, h1)

		_, _ = fmt.Fprintf(w, fmProj, p.ProjectName)
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
		_, _ = fmt.Fprintf(
			w, fmName, utils.Ellipsis(epic.MySummary(), p.FieldSizeName))
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprint(
			w, dr.AsIntersect(
				today, p.Outer, p.UseColor,
				myj.StatusColor(epic.Status(), myj.ColorKindTerminal)))
		_, _ = fmt.Fprintln(w)
		lineCount++
		if lineCount%p.LineSetSize == 0 {
			_, _ = fmt.Fprintln(w)
		}
	}
}
