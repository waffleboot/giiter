package git

import (
	"errors"
)

type Message struct {
	Subject     string
	Description string
}

type commit struct {
	SHA     string
	Message Message
}

type Branch struct {
	CommitSHA  string
	BranchName string
}

type reviewBranch struct {
	id     int
	branch Branch
}

type Record struct {
	NewID          int
	FeatureSHA     string
	FeatureMsg     Message
	reviewBranches reviewBranches
}

type reviewBranches struct {
	commitSHA string
	reviewMsg Message
	branches  []reviewBranch
}

func (r *Record) HasReview() bool {
	return r.reviewBranches.commitSHA != ""
}

func (r *Record) IsNewCommit() bool {
	return r.reviewBranches.commitSHA == ""
}

func (r *Record) IsOldCommit() bool {
	return r.FeatureSHA == ""
}

func (r *Record) MatchedCommit() bool {
	return r.FeatureSHA == r.reviewBranches.commitSHA
}

func (r *reviewBranches) AddReviewBranch(branch reviewBranch) {
	r.commitSHA = branch.branch.CommitSHA
	r.branches = append(r.branches, branch)
}

func (r *reviewBranches) MaxID() int {
	var maxID int
	for _, branch := range r.branches {
		if branch.id > maxID {
			maxID = branch.id
		}
	}

	return maxID
}

func (r *Record) ReviewBranchNames() []string {
	return r.reviewBranches.reviewBranchNames()
}

func (r *Record) AnyReviewBranch() (string, error) {
	return r.reviewBranches.anyReviewBranch()
}

func (r *Record) CommitSHA() string {
	if r.FeatureSHA != "" {
		return r.FeatureSHA
	}

	return r.reviewBranches.commitSHA
}

func (r *Record) CommitMessage() Message {
	if r.FeatureSHA != "" {
		return r.FeatureMsg
	}

	return r.reviewBranches.reviewMsg
}

func (r *Record) switchBranch() {
	r.reviewBranches.commitSHA = r.FeatureSHA
	r.reviewBranches.reviewMsg = r.FeatureMsg
}

func (r *reviewBranches) reviewBranchNames() []string {
	a := make([]string, 0, len(r.branches))
	for _, branch := range r.branches {
		a = append(a, branch.BranchName())
	}

	return a
}

func (r *reviewBranches) anyReviewBranch() (string, error) {
	if len(r.branches) > 1 {
		return "", errors.New("unable to choose any review branch")
	}

	return r.branches[0].BranchName(), nil
}

func (r *reviewBranch) BranchName() string {
	return r.branch.BranchName
}
