package git

type Message struct {
	Subject     string
	Description string
}

type Commit struct {
	SHA     string
	Message Message
}

type Branch struct {
	CommitSHA  string
	BranchName string
}

type ReviewBranch struct {
	ID int
	Branch
}

type Record struct {
	NewID      int
	FeatureSHA string
	FeatureMsg Message
	ReviewBranches
}

type ReviewBranches struct {
	CommitSHA      string
	ReviewMsg      Message
	ReviewBranches []ReviewBranch
}

func (r *Record) HasReview() bool {
	return r.ReviewBranches.CommitSHA != ""
}

func (r *Record) IsNewCommit() bool {
	return r.ReviewBranches.CommitSHA == ""
}

func (r *Record) IsOldCommit() bool {
	return r.FeatureSHA == ""
}

func (r *Record) MatchedCommit() bool {
	return r.FeatureSHA == r.ReviewBranches.CommitSHA
}

func (r *ReviewBranches) AddReviewBranch(branch ReviewBranch) {
	r.CommitSHA = branch.CommitSHA
	r.ReviewBranches = append(r.ReviewBranches, branch)
}

func (r *ReviewBranches) MaxID() int {
	var maxID int
	for _, branch := range r.ReviewBranches {
		if branch.ID > maxID {
			maxID = branch.ID
		}
	}
	return maxID
}

func (r *ReviewBranches) ReviewBranchNames() []string {
	a := make([]string, 0, len(r.ReviewBranches))
	for _, branch := range r.ReviewBranches {
		a = append(a, branch.BranchName)
	}
	return a
}

// func (r *ReviewBranches) MainReviewBranch() string {
// 	return r.ReviewBranches[0].BranchName
// }
