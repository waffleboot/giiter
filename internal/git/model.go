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
	return r.FeatureSHA == r.ReviewBranch.CommitSHA
}
