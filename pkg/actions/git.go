package actions

import (
	"bufio"
	"fmt"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"os"
	"path/filepath"
	"strings"
)

func UpdateGitignore(c *config.Config) error {
	path := filepath.Join(c.Root, ".gitignore")
	ignores := c.Suffixes.GitignoreSet()
	ignores["/"+cache.cacheDirName] = true
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
