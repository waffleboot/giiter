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
	featureSHA     string
	featureMsg     Message
	reviewSHA      string
	reviewMsg      Message
	reviewBranches reviewBranches
}

type reviewBranches struct {
	branches []reviewBranch
}

func (r *Record) HasReview() bool {
	return r.reviewSHA != ""
}

func (r *Record) IsNewCommit() bool {
	return r.reviewSHA == ""
}

func (r *Record) IsOldCommit() bool {
	return r.featureSHA == ""
}

func (r *Record) MatchedCommit() bool {
	return r.featureSHA == r.reviewSHA
}

func (r *Record) addReviewBranch(branch reviewBranch) {
	r.reviewSHA = branch.branch.CommitSHA
	r.reviewBranches.addReviewBranch(branch)
}

func (r *Record) ReviewBranchNames() []string {
	return r.reviewBranches.reviewBranchNames()
}

func (r *Record) AnyReviewBranch() (string, error) {
	return r.reviewBranches.anyReviewBranch()
}

func (r *Record) CommitSHA() string {
	if r.featureSHA != "" {
		return r.featureSHA
	}

	return r.reviewSHA
}

func (r *Record) CommitMessage() Message {
	if r.featureSHA != "" {
		return r.featureMsg
	}

	return r.reviewMsg
}

func (r *Record) switchBranch() {
	r.reviewSHA = r.featureSHA
	r.reviewMsg = r.featureMsg
}

func newRecord(commit *commit) Record {
	return Record{
		featureSHA: commit.SHA,
		featureMsg: commit.Message,
	}
}

func newReviewRecord(commit *commit, branch reviewBranch) Record {
	return Record{
		reviewSHA: commit.SHA,
		reviewMsg: commit.Message,
		reviewBranches: reviewBranches{
			branches: []reviewBranch{branch},
		},
	}
}

func newReviewBranch(id int, branch Branch) reviewBranch {
	return reviewBranch{
		id:     id,
		branch: branch,
	}
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

func (r *reviewBranches) addReviewBranch(branch reviewBranch) {
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
