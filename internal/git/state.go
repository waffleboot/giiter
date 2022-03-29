package git

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/waffleboot/giiter/internal/config"
)

func (g *git) State(base, feat string) ([]Record, error) {
	commits, err := g.Commits(base, feat)
	if err != nil {
		return nil, err
	}

	records := make([]Record, 0, len(commits))

	featureSHAIndex := make(map[string]int)

	featureSubjIndex := make(map[string]int)

	for i := range commits {

		commit, err := g.FindCommit(commits[i])
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

	branches, err := g.Branches()
	if err != nil {
		return nil, err
	}

	reviewBranchPrefix := fmt.Sprintf("review/%s/", feat)

	var featureDiffHashToIndex map[string]int

	for i := range branches {
		branch := branches[i]
		if !strings.HasPrefix(branch.Name, reviewBranchPrefix) {
			continue
		}

		reviewSHA := branch.SHA

		branchSuffix := branch.Name[len(reviewBranchPrefix):]

		id, err := strconv.Atoi(branchSuffix)
		if err != nil {
			return nil, err
		}

		if index, ok := featureSHAIndex[reviewSHA]; ok {
			records[index].ID = id
			records[index].ReviewSHA = reviewSHA
			records[index].ReviewMsg = records[index].FeatureMsg
			records[index].ReviewBranch = branch.Name
			continue
		}

		if featureDiffHashToIndex == nil {

			featureDiffHashToIndex = make(map[string]int)

			for i := range records {
				diffHash, err := g.DiffHash(records[i].FeatureSHA)
				if err != nil {
					return nil, err
				}
				if diffHash.valid {
					featureDiffHashToIndex[diffHash.hash] = i
				}
			}

		}

		diffHash, err := g.DiffHash(reviewSHA)
		if err != nil {
			return nil, err
		}

		commit, err := g.FindCommit(reviewSHA)
		if err != nil {
			return nil, err
		}

		if diffHash.valid {
			if index, ok := featureDiffHashToIndex[diffHash.hash]; ok {
				records[index].ID = id
				records[index].ReviewSHA = commit.SHA
				records[index].ReviewMsg = commit.Message
				records[index].ReviewBranch = branch.Name
				continue
			}
		}

		if config.Config.RefreshOnSubject {
			if index, ok := featureSubjIndex[commit.Message.Subject]; ok {
				records[index].ID = id
				records[index].ReviewSHA = commit.SHA
				records[index].ReviewMsg = commit.Message
				records[index].ReviewBranch = branch.Name
				continue
			}
		}

		record := Record{
			ID:           id,
			FeatureSHA:   "",
			FeatureMsg:   Message{"", ""},
			ReviewSHA:    commit.SHA,
			ReviewMsg:    commit.Message,
			ReviewBranch: branch.Name,
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

func (g *git) Refresh(base, feat string) ([]Record, error) {
	records, err := g.State(base, feat)
	if err != nil {
		return nil, err
	}

	for i := range records {
		record := records[i]
		if record.IsOldCommit() || record.IsNewCommit() || record.FeatureSHA == record.ReviewSHA {
			continue
		}
		if err = g.SwitchBranch(record.ReviewBranch, record.FeatureSHA); err != nil {
			return nil, err
		}
		records[i].ReviewSHA = record.FeatureSHA
		records[i].ReviewMsg = record.FeatureMsg
	}

	for i := range records {
		if records[i].IsNewCommit() {
			return records, nil
		}
	}

	newRecords := make([]Record, 0, len(records))
	for i := range records {
		if records[i].FeatureSHA != "" {
			newRecords = append(newRecords, records[i])
			continue
		}
		if err := g.DeleteBranch(records[i].ReviewBranch); err != nil {
			return nil, err
		}
	}

	return newRecords, nil
}
