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
	NewID          int
	FeatureSHA     string
	FeatureMsg     Message
	ReviewMsg      Message
	ReviewBranches []ReviewBranch
}

func (r *Record) HasReview() bool {
	return r.ReviewBranches != nil
}

func (r *Record) IsNewCommit() bool {
	return r.ReviewBranches == nil
}

func (r *Record) IsOldCommit() bool {
	return r.FeatureSHA == ""
}

func (r *Record) MatchedCommit() bool {
	return r.FeatureSHA == r.ReviewBranches[0].CommitSHA
}

func (r *Record) AddReviewBranch(branch ReviewBranch) {
	r.ReviewBranches = append(r.ReviewBranches, branch)
}

func (r *Record) MaxID() int {
	var maxID int
	for _, branch := range r.ReviewBranches {
		if branch.ID > maxID {
			maxID = branch.ID
		}
	}
	return maxID
}
