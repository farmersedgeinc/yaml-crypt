package cmd

import (
	"os"
	"os/exec"

	"github.com/farmersedgeinc/yaml-crypt/pkg/actions"
	"github.com/farmersedgeinc/yaml-crypt/pkg/cache"
	"github.com/farmersedgeinc/yaml-crypt/pkg/config"
	"github.com/google/shlex"
	"github.com/spf13/cobra"
)

var editFlags struct {
	editor string
}

var editCmd = &cobra.Command{
	Use:                   "edit <file>",
	Short:                 "edit a file in your $EDITOR",
	Long:                  "edit a file in your $EDITOR. The equivalent of running \"yaml-crypt decrypt <file> && $EDITOR <file> && yaml-crypt encrypt <file>\".",
	Args:                  cobra.ExactArgs(1),
	DisableFlagsInUseLine: true,
	RunE: func(_ *cobra.Command, args []string) error {
		var err error

		// figure out editor
		var editor string
		if e := editFlags.editor; e != "$EDITOR" {
			editor = editFlags.editor
		} else if e := os.Getenv("EDITOR"); e != "" {
			editor = e
		} else {
			editor = "vi"
		}
		editorFlags := []string{}
		if f := os.Getenv("EDITORFLAGS"); f != "" {
			editorFlags, err = shlex.Split(f)
			if err != nil {
				return err
			}
		}

		// get file
		config, err := config.LoadConfig(".")
		if err != nil {
			return err
		}
		file, err := actions.NewFile(args[0], &config)
		if err != nil {
			return err
		}

		// decrypt
		err = func() error {
			cache, err := cache.Setup(config)
			if err != nil {
				return err
			}
			defer cache.Close()
			return actions.Decrypt([]*actions.File{&file}, false, false, &cache, &config.Provider, int(threads), retries, timeout, progress)
		}()
		if err != nil {
			return err
		}

		// edit
		editorFlags = append(editorFlags, file.DecryptedPath)
		cmd := exec.Command(editor, editorFlags...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err = cmd.Run()
		if err != nil {
			return err
		}

		// re-open cache
		cache, err := cache.Setup(config)
		if err != nil {
			return err
		}
		defer cache.Close()

		// encrypt
		err = actions.Encrypt([]*actions.File{&file}, &cache, &config.Provider, int(threads), retries, timeout, progress)
		if err != nil {
			return err
		}

		// cleanup decrypted file
		err = os.Remove(file.DecryptedPath)
		if err != nil {
			return err
		}

		// if plain file doesn't exist, we're done
		if _, err := os.Stat(file.PlainPath); os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return err
		}
		// update plain file
		return actions.Decrypt([]*actions.File{&file}, true, false, &cache, &config.Provider, int(threads), retries, timeout, progress)
	},
}

func init() {
	rootCmd.AddCommand(editCmd)
	editCmd.Flags().StringVarP(&editFlags.editor, "editor", "e", "$EDITOR", "editor to use")
}
