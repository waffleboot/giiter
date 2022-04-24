package git

import "context"

func Refresh(ctx context.Context, baseBranch, featureBranch string) ([]Record, error) {
	records, err := State(ctx, baseBranch, featureBranch)
	if err != nil {
		return nil, err
	}

	// переключить feature коммиты на найденные review ветки

	for i := range records {
		record := records[i]
		if record.IsOldCommit() || record.IsNewCommit() || record.MatchedCommit() {
			continue
		}

		if errSwitch := SwitchBranch(ctx, record.MainReviewBranch(), record.FeatureSHA); errSwitch != nil {
			return nil, errSwitch
		}

		records[i].ReviewBranches.CommitSHA = record.FeatureSHA
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

		for _, branch := range records[i].ReviewBranchNames() {
			if err := DeleteBranch(ctx, branch); err != nil {
				return nil, err
			}
		}
	}

	return newRecords, nil
}
