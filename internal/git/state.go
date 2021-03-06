package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/waffleboot/giiter/internal/app"
)

const (
	Prefix = "review/"
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
		commit, errFind := findCommit(ctx, commits[i])
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
		review := branches[i]

		reviewSHA := review.branch.CommitSHA
		if index, ok := r.shaIndex[reviewSHA]; ok {
			r.records[index].addReviewBranch(review)

			continue
		}

		if errLazy := r.lazyDiffHashes(ctx); errLazy != nil {
			return nil, errLazy
		}

		commit, err := findCommit(ctx, reviewSHA)
		if err != nil {
			return nil, err
		}

		diffHash, err := diffHash(ctx, reviewSHA)
		if err != nil {
			return nil, err
		}

		if diffHash.Valid {
			if index, ok := r.diffIndex[diffHash.String]; ok {
				r.records[index].addReviewBranch(review)

				continue
			}
		}

		if app.Config.UseSubjectToMatch {
			if index, ok := r.subjIndex[commit.Message.Subject]; ok {
				r.records[index].addReviewBranch(review)

				continue
			}
		}

		r.addReviewRecord(review, commit)
	}

	r.fillNewCommitIDs()

	return r.records, nil
}

func (r *records) addReviewRecord(branch reviewBranch, commit *commit) {
	r.shaIndex[commit.SHA] = len(r.records)

	r.records = append(r.records, newReviewRecord(commit, branch))
}

func AllReviewBranches(ctx context.Context, featureBranch string) (result []reviewBranch, err error) {
	branchPrefix := fmt.Sprintf(Prefix+"%s/", featureBranch)

	branches, err := AllBranches(ctx, runner{})
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

			if diffHash.Valid {
				r.diffIndex[diffHash.String] = i
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
