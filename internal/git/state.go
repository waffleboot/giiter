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

		commit, err := FindCommit(ctx, commits[i])
		if err != nil {
			return nil, err
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
				diffHash, err := DiffHash(ctx, records[i].FeatureSHA)
				if err != nil {
					return nil, err
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

		diffHash, err := DiffHash(ctx, reviewSHA)
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

func Refresh(ctx context.Context, baseBranch, featureBranch string) ([]Record, error) {
	records, err := State(ctx, baseBranch, featureBranch)
	if err != nil {
		return nil, err
	}

	// переключить feature коммиты на найденные review ветки

	for i := range records {
		record := records[i]
		if record.IsOldCommit() || record.IsNewCommit() || record.FeatureSHA == record.ReviewSHA {
			continue
		}
		if err = SwitchBranch(ctx, record.ReviewBranch, record.FeatureSHA); err != nil {
			return nil, err
		}
		records[i].ReviewSHA = record.FeatureSHA
		records[i].ReviewMsg = record.FeatureMsg
	}

	// если хотя бы один новый коммит не сопоставленный остался, то заброшенные review ветки не удаляем
	// чтобы можно было сделать ручной assign коммитов на эти ветки, чтобы не потерять review comments

	for i := range records {
		if records[i].IsNewCommit() {
			return records, nil
		}
	}

	// так как все коммиты на своих review ветках, можно удалять старые review ветки
	// коммиты на них устарели

	newRecords := make([]Record, 0, len(records))
	for i := range records {
		if records[i].FeatureSHA != "" {
			newRecords = append(newRecords, records[i])
			continue
		}
		if err := DeleteBranch(ctx, records[i].ReviewBranch); err != nil {
			return nil, err
		}
	}

	return newRecords, nil
}
