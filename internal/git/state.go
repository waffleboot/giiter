package git

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/waffleboot/giiter/internal/app"
)

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

	branches, err := AllBranches(ctx)
	if err != nil {
		return nil, err
	}

	reviewBranchPrefix := fmt.Sprintf("review/%s/", featureBranch)

	var featureDiffToIndex map[string]int

	for i := range branches {
		reviewBranch := branches[i]
		if !strings.HasPrefix(reviewBranch.BranchName, reviewBranchPrefix) {
			continue
		}

		reviewSHA := reviewBranch.CommitSHA

		branchSuffix := reviewBranch.BranchName[len(reviewBranchPrefix):]

		id, err := strconv.Atoi(branchSuffix)
		if err != nil {
			return nil, err
		}

		if index, ok := featureSHAIndex[reviewSHA]; ok {
			records[index].ID = id
			records[index].ReviewSHA = reviewSHA
			records[index].ReviewMsg = records[index].FeatureMsg
			records[index].ReviewBranch = reviewBranch.BranchName

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
				records[index].ID = id
				records[index].ReviewSHA = commit.SHA
				records[index].ReviewMsg = commit.Message
				records[index].ReviewBranch = reviewBranch.BranchName

				continue
			}
		}

		if app.Config.UseSubjectToMatch {
			if index, ok := featureSubjIndex[commit.Message.Subject]; ok {
				records[index].ID = id
				records[index].ReviewSHA = commit.SHA
				records[index].ReviewMsg = commit.Message
				records[index].ReviewBranch = reviewBranch.BranchName

				continue
			}
		}

		record := Record{
			ID:           id,
			FeatureSHA:   "",
			FeatureMsg:   Message{"", ""},
			ReviewSHA:    commit.SHA,
			ReviewMsg:    commit.Message,
			ReviewBranch: reviewBranch.BranchName,
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

