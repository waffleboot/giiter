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
	NewID        int
	FeatureSHA   string
	FeatureMsg   Message
	ReviewMsg    Message
	ReviewBranch ReviewBranch
}

func (r *Record) HasReview() bool {
	return r.ReviewBranch.CommitSHA != ""
}

func (r *Record) IsNewCommit() bool {
	return r.ReviewBranch.CommitSHA == ""
}

func (r *Record) IsOldCommit() bool {
	return r.FeatureSHA == ""
}

func (r *Record) MatchedCommit() bool {
	return r.FeatureSHA == r.ReviewBranch.CommitSHA
}
