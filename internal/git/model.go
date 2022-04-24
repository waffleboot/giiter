package git

import (
	"errors"
)

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
	id     int
	branch Branch
}

type Record struct {
	NewID          int
	FeatureSHA     string
	FeatureMsg     Message
	ReviewBranches ReviewBranches
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
	r.CommitSHA = branch.branch.CommitSHA
	r.ReviewBranches = append(r.ReviewBranches, branch)
}

func (r *ReviewBranches) MaxID() int {
	var maxID int
	for _, branch := range r.ReviewBranches {
		if branch.id > maxID {
			maxID = branch.id
		}
	}

	return maxID
}

func (r *ReviewBranches) ReviewBranchNames() []string {
	a := make([]string, 0, len(r.ReviewBranches))
	for _, branch := range r.ReviewBranches {
		a = append(a, branch.BranchName())
	}

	return a
}

func (r *ReviewBranches) AnyReviewBranch() (string, error) {
	if len(r.ReviewBranches) > 1 {
		return "", errors.New("unable to choose any review branch")
	}

	return r.ReviewBranches[0].BranchName(), nil
}

func (r *ReviewBranch) BranchName() string {
	return r.branch.BranchName
}
