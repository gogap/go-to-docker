package builder

import (
	"io"
	"os"
	"strings"
)

func copyfile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}

	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return
	}

	defer func() {
		cerr := out.Close()
		if err == nil {
			err = cerr
		}
	}()

	if _, err = io.Copy(out, in); err != nil {
		return
	}
	err = out.Sync()

	return
}

func getRevision(dir string) (branchName, revisionID string, isGit bool, err error) {
	cmdRevID := `git rev-parse HEAD`
	cmdRevBranch := `git rev-parse --abbrev-ref HEAD`
	cmdIsGit := `git rev-parse --git-dir`
	if _, e := execCommand(dir, cmdIsGit); e != nil {
		return
	}

	var out1, out2 []byte
	if out1, err = execCommand(dir, cmdRevBranch); err != nil {
		if strings.Contains(string(out1), "HEAD") {
			err = nil
			out1 = []byte("master")
		} else {
			return
		}
	}

	if out2, err = execCommand(dir, cmdRevID); err != nil {
		if strings.Contains(string(out2), "HEAD") {
			err = nil
			out2 = []byte("0000000000000000000000000000000000000000")
		} else {
			return
		}
	}

	branchName = strings.Trim(string(out1), "\n")
	revisionID = string(out2[0:8])
	isGit = true

	return
}
