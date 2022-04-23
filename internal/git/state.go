package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/waffleboot/giiter/internal/app"
)

func findReviewBranches(ctx context.Context, featureBranch string) ([]ReviewBranch, error) {
	branches, err := AllBranches(ctx)
	if err != nil {
		return nil, err
	}

	branchPrefix := fmt.Sprintf("review/%s/", featureBranch)

	reviewBranches := make([]ReviewBranch, 0, 16)

	for _, branch := range branches {
		if !strings.HasPrefix(branch.BranchName, branchPrefix) {
			continue
		}

		branchSuffix := branch.BranchName[len(branchPrefix):]

		id, err := strconv.Atoi(branchSuffix)
		if err != nil {
			return nil, err
		}

		reviewBranches = append(reviewBranches, ReviewBranch{
			ID:     id,
			Branch: branch,
		})
	}

	return reviewBranches, nil
}

func State(ctx context.Context, baseBranch, featureBranch string) ([]Record, error) {
	commits, err := Commits(ctx, baseBranch, featureBranch)
	if err != nil {
		return nil, errors.WithMessage(err, "get state")
	}

	records := make([]Record, 0, len(commits))

	featureSHAIndex := make(map[string]int)

	featureSubjIndex := make(map[string]int)

	for i := range commits {
		commit, errCommit := FindCommit(ctx, commits[i])
		if errCommit != nil {
			return nil, errCommit
		}

		record := Record{
			FeatureSHA: commit.SHA,
			FeatureMsg: commit.Message,
		}
		records = append(records, record)

		featureSHAIndex[commit.SHA] = i

		featureSubjIndex[commit.Message.Subject] = i
	}

	branches, err := findReviewBranches(ctx, featureBranch)
	if err != nil {
		return nil, err
	}

	var featureDiffToIndex map[string]int

	for i := range branches {
		branch := branches[i]

		reviewSHA := branch.CommitSHA
		if index, ok := featureSHAIndex[reviewSHA]; ok {
			records[index].ID = branch.ID
			records[index].ReviewSHA = reviewSHA
			records[index].ReviewMsg = records[index].FeatureMsg
			records[index].ReviewBranch = branch.BranchName

			continue
		}

		if featureDiffToIndex == nil {
			featureDiffToIndex = make(map[string]int)

			for i := range records {
				diffHash, errDiff := diffHash(ctx, records[i].FeatureSHA)
				if errDiff != nil {
					return nil, errDiff
				}

				if diffHash.valid {
					featureDiffToIndex[diffHash.hash] = i
				}
			}
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
			if index, ok := featureDiffToIndex[diffHash.hash]; ok {
				records[index].ID = branch.ID
				records[index].ReviewSHA = commit.SHA
				records[index].ReviewMsg = commit.Message
				records[index].ReviewBranch = branch.BranchName

				continue
			}
		}

		if app.Config.UseSubjectToMatch {
			if index, ok := featureSubjIndex[commit.Message.Subject]; ok {
				records[index].ID = branch.ID
				records[index].ReviewSHA = commit.SHA
				records[index].ReviewMsg = commit.Message
				records[index].ReviewBranch = branch.BranchName

				continue
			}
		}

		record := Record{
			ID:           branch.ID,
			FeatureSHA:   "",
			FeatureMsg:   Message{"", ""},
			ReviewSHA:    commit.SHA,
			ReviewMsg:    commit.Message,
			ReviewBranch: branch.BranchName,
		}

		records = append(records, record)
	}

	var maxID int

	for i := range records {
		if records[i].ID > maxID {
			maxID = records[i].ID
		}
	}

	for i := range records {
		if records[i].IsNewCommit() {
			maxID++
			records[i].ID = maxID
		}
	}

	return records, nil
}

