package actions

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/farmersedgeinc/yaml-crypt/pkg/cache/disk"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
)

func UpdateGitignore(c *config.Config) error {
	path := filepath.Join(c.Root, ".gitignore")
	ignores := c.Suffixes.GitignoreSet()
	ignores["/"+disk.CacheDirName] = true
	if exists(path) {
		existingFile, err := os.Open(path)
		defer existingFile.Close()
		if err != nil {
			return err
		}
		tmpFilename := path + ".tmp"
		newFile, err := os.Create(tmpFilename)
		newFile.Chmod(0644)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(existingFile)
		for scanner.Scan() {
			line := strings.Trim(scanner.Text(), "\r\n")
			if _, ok := ignores[line]; ok {
				delete(ignores, line)
			}
			_, err = fmt.Fprintln(newFile, line)
			if err != nil {
				return err
			}
		}
		if err = scanner.Err(); err != nil {
			return err
		}
		for ignore := range ignores {
			_, err = fmt.Fprintln(newFile, ignore)
			if err != nil {
				return err
			}
		}
		err = existingFile.Close()
		if err != nil {
			return err
		}
		err = newFile.Close()
		if err != nil {
			return err
		}
		return os.Rename(tmpFilename, path)
	} else {
		newFile, err := os.Create(path)
		newFile.Chmod(0644)
		if err != nil {
			return err
		}
		for ignore := range ignores {
			_, err := fmt.Fprintln(newFile, ignore)
			if err != nil {
				return err
			}
		}
		err = newFile.Close()
		return err
	}
}
