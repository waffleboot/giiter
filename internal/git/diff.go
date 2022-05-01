package git

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/waffleboot/giiter/internal/app"
)

func diffFiles(ctx context.Context, sha string) ([]string, error) {
	files, err := run(ctx, "diff-tree", "-r", "--name-only", "-c", sha)
	if err != nil {
		return nil, err
	}
	if len(files) == 1 {
		return nil, nil
	}
	return files[1:], nil
}

func diffHash(ctx context.Context, sha string) (sql.NullString, error) {
	files, err := diffFiles(ctx, sha)
	if err != nil {
		return sql.NullString{}, err
	}

	hash := sha256.New()

	for i := range files {
		file := files[i]

		diff, err := run(ctx, "diff-tree", "--unified=0", sha, "--", file)
		if err != nil {
			return sql.NullString{}, err
		}

		diff = diff[2:]

		if strings.HasPrefix(diff[0], "new file") {
			hash.Write([]byte("new file"))
			hash.Write([]byte(file))

			continue
		}

		if strings.HasPrefix(diff[0], "deleted file") {
			hash.Write([]byte("deleted file"))
			hash.Write([]byte(file))

			continue
		}

		if strings.HasPrefix(diff[1], "Binary") {
			return sql.NullString{}, nil
		}

		diff = diff[3:]

		for _, line := range diff {
			hash.Write([]byte(line))
		}

		if app.Config.Debug {
			fmt.Printf("--- diff %s %s\n", sha, file)

			for _, line := range diff {
				fmt.Println(line)
			}
		}
	}

	sum := hash.Sum(nil)

	strSum := fmt.Sprintf("%x", sum)

	if app.Config.Debug {
		fmt.Printf("Commit: %s DiffHash: %s\n", sha, strSum)
	}

	return sql.NullString{String: strSum, Valid: true}, nil
}

func Diff(ctx context.Context, commitSHA string, args ...string) error {
	cmdArgs := []string{"diff", commitSHA + "~.." + commitSHA}

	cmd := exec.CommandContext(ctx, "git", append(cmdArgs, args...)...)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout

	return cmd.Run()
}
