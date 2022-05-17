package git

import (
	"context"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"

	"github.com/waffleboot/giiter/internal/git/mocks"
)

func TestAllBranches(t *testing.T) {
	mc := minimock.NewController(t)
	mo := mocks.NewGitRunnerMock(mc).AllBranchesMock.Return([]string{
		"1234567 branch",
		"12345678 branch",
	}, nil)
	branches, err := AllBranches(context.Background(), mo)
	require.NoError(t, err)
	require.Equal(t, []Branch{
		{
			CommitSHA:  "1234567",
			BranchName: "branch",
		},
		{
			CommitSHA:  "12345678",
			BranchName: "branch",
		},
	}, branches)
}

func TestChangedFiles(t *testing.T) {
	mc := minimock.NewController(t)
	mo := mocks.NewGitRunnerMock(mc).ChangedFilesMock.Return([]string{
		"bd291ea7324d5812eefcf3fa17b307b2d0f30660",
		"\"\\321\\200\\321\\203\\321\\201\\321\\201\\320\\272\\320\\270\\320\\271\"",
	}, nil)
	files, err := changedFiles(context.Background(), "sha", mo)
	require.NoError(t, err)
	require.Equal(t, []string{"русский"}, files)
}
