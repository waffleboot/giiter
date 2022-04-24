package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/waffleboot/giiter/internal/app"
)

type records struct {
	records   []Record
	shaIndex  map[string]int
	subjIndex map[string]int
	diffIndex map[string]int
}

func State(ctx context.Context, baseBranch, featureBranch string) ([]Record, error) {
	r, err := createRecords(ctx, baseBranch, featureBranch)
	if err != nil {
		return nil, err
	}

	branches, err := AllReviewBranches(ctx, featureBranch)
	if err != nil {
		return nil, err
	}

	return r.matchCommitsAndBranches(ctx, branches)
}

func createRecords(ctx context.Context, baseBranch, featureBranch string) (*records, error) {
	commits, err := Commits(ctx, baseBranch, featureBranch)
	if err != nil {
		return nil, errors.WithMessage(err, "get state")
	}

	r := &records{
		records:   make([]Record, 0, len(commits)),
		shaIndex:  make(map[string]int),
		subjIndex: make(map[string]int),
	}

	for i := range commits {
		commit, errFind := FindCommit(ctx, commits[i])
		if errFind != nil {
			return nil, errFind
		}

		r.records = append(r.records, newRecord(commit))

		r.shaIndex[commit.SHA] = i

		r.subjIndex[commit.Message.Subject] = i
	}

	return r, nil
}

func (r *records) matchCommitsAndBranches(ctx context.Context, branches []reviewBranch) ([]Record, error) {
	for i := range branches {
		branch := branches[i]

		reviewSHA := branch.branch.CommitSHA
		if index, ok := r.shaIndex[reviewSHA]; ok {
			r.records[index].addReviewBranch(branch)

			continue
		}

		if errLazy := r.lazyDiffHashes(ctx); errLazy != nil {
			return nil, errLazy
		}

		commit, err := FindCommit(ctx, reviewSHA)
		if err != nil {
			return nil, err
		}

		diffHash, err := diffHash(ctx, reviewSHA)
		if err != nil {
			return nil, err
		}

		if diffHash.valid {
			if index, ok := r.diffIndex[diffHash.hash]; ok {
				r.records[index].addReviewBranch(branch)

				continue
			}
		}

		if app.Config.UseSubjectToMatch {
			if index, ok := r.subjIndex[commit.Message.Subject]; ok {
				r.records[index].addReviewBranch(branch)

				continue
			}
		}

		r.addReviewRecord(branch, commit)
	}

	r.fillNewCommitIDs()

	return r.records, nil
}

func (r *records) addReviewRecord(branch reviewBranch, commit *commit) {
	r.shaIndex[commit.SHA] = len(r.records)

	r.records = append(r.records, newReviewRecord(commit, branch))
}

func AllReviewBranches(ctx context.Context, featureBranch string) (result []reviewBranch, err error) {
	branchPrefix := fmt.Sprintf("review/%s/", featureBranch)

	branches, err := AllBranches(ctx)
	if err != nil {
		return nil, err
	}

	for _, branch := range branches {
		if !strings.HasPrefix(branch.BranchName, branchPrefix) {
			continue
		}

		branchSuffix := branch.BranchName[len(branchPrefix):]

		id, err := strconv.Atoi(branchSuffix)
		if err != nil {
			return nil, err
		}

		result = append(result, newReviewBranch(id, branch))
	}

	return
}

func (r *records) lazyDiffHashes(ctx context.Context) error {
	if r.diffIndex == nil {
		r.diffIndex = make(map[string]int)

		for i := range r.records {
			diffHash, err := diffHash(ctx, r.records[i].CommitSHA())
			if err != nil {
				return err
			}

			if diffHash.valid {
				r.diffIndex[diffHash.hash] = i
			}
		}
	}

	return nil
}

func (r *records) fillNewCommitIDs() {
	var maxID int

	for i := range r.records {
		if r.records[i].MaxID() > maxID {
			maxID = r.records[i].MaxID()
		}
	}

	for i := range r.records {
		if r.records[i].IsNewCommit() {
			maxID++
			r.records[i].NewID = maxID
		}
	}
}
