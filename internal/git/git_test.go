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
