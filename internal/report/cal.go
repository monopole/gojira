package report

import (
	"fmt"
	"github.com/monopole/gojira/internal/myj"
	"github.com/monopole/gojira/internal/utils"
	"io"
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
	ShowAssignee  bool
}

func DoCal(
	w io.Writer,
	epicMap map[myj.MyKey]*myj.ResponseIssue,
	p CalParams,
) error {
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
	var errors []error
	for _, epicKey := range epicKeys {
		epic := epicMap[epicKey.MyKey]
		dr, err := utils.MakeDayRangeGentle(epic.DateStart(), epic.DateEnd())
		if err != nil {
			errors = append(
				errors,
				fmt.Errorf("%s; %w", epicKey.MyKey, err))
		}
		_, _ = fmt.Fprintf(w, fmId, epic.MyKey.Num)
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprintf(
			w, fmName, utils.Ellipsis(epic.MySummary(), p.FieldSizeName))
		_, _ = fmt.Fprint(w, spacer)
		_, _ = fmt.Fprint(
			w, dr.AsIntersect(
				today,
				func() string {
					if p.ShowAssignee {
						return epic.AssigneeName()
					}
					return ""
				}(),
				p.Outer, p.UseColor,
				myj.StatusColor(epic.Status(), myj.ColorKindTerminal)))
		_, _ = fmt.Fprintln(w)
		lineCount++
		if lineCount%p.LineSetSize == 0 {
			_, _ = fmt.Fprintln(w)
		}
	}
	if len(errors) == 0 {
		return nil
	}
	return fmt.Errorf("detected %d date errors, e.g. %w", len(errors), errors[0])
}
