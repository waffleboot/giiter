package git

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/waffleboot/giiter/internal/app"
)

func changedFiles(ctx context.Context, sha string, runner Runner) ([]string, error) {
	files, err := runner.ChangedFiles(ctx, sha)
	if err != nil {
		return nil, err
	}

	if len(files) == 1 {
		return nil, nil
	}

	files = files[1:]
	for i := range files {
		if files[i][0] == '"' {
			files[i], err = strconv.Unquote(files[i])
			if err != nil {
				return nil, err
			}
		}
	}

	return files, nil
}

func diffHash(ctx context.Context, sha string) (sql.NullString, error) {
	files, err := changedFiles(ctx, sha, runner{})
	if err != nil {
		return sql.NullString{}, err
	}

	if len(files) == 0 {
		return sql.NullString{}, nil
	}

	hash := sha256.New()

	for i := range files {
		file := files[i]

		diff, err := run(ctx, "diff-tree", "--unified=0", "-c", sha, "--", file)
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
