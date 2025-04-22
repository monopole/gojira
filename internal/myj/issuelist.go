package myj

type IssueList []*ResponseIssue

func (mi IssueList) Len() int {
	return len(mi)
}

func (mi IssueList) Less(i, j int) bool {
	if mi[i].MyKey.Proj == mi[j].MyKey.Proj {
		return mi[i].MyKey.Num < mi[j].MyKey.Num
	}
	return mi[i].MyKey.Proj < mi[j].MyKey.Proj
}

func (mi IssueList) Swap(i, j int) {
	mi[i], mi[j] = mi[j], mi[i]
}
