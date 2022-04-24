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

		for _, branch := range record.ReviewBranchNames() {
			if errSwitch := SwitchBranch(ctx, branch, record.featureSHA); errSwitch != nil {
				return nil, errSwitch
			}
		}

		records[i].switchBranch()
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

	var j int

	for _, record := range records {
		if record.IsOldCommit() {
			if err := deleteReviewBranches(ctx, record); err != nil {
				return nil, err
			}

			continue
		}

		records[j] = record
		j++
	}

	return records[:j], nil
}

func deleteReviewBranches(ctx context.Context, record Record) error {
	for _, branch := range record.ReviewBranchNames() {
		if err := DeleteBranch(ctx, branch); err != nil {
			return err
		}
	}

	return nil
}
