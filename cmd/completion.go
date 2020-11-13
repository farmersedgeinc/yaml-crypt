package cmd

import (
	"github.com/spf13/cobra"
	"os"
)

var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "generate completion script",
	Long: `To load completions:
Bash:

$ source <(yaml-crypt completion bash)

# To load completions for each session, execute once:
Linux:
  $ yaml-crypt completion bash > /etc/bash_completion.d/yaml-crypt
MacOS:
  $ yaml-crypt completion bash > /usr/local/etc/bash_completion.d/yaml-crypt

Zsh:

# If shell completion is not already enabled in your environment you will need
# to enable it.  You can execute the following once:

$ echo "autoload -U compinit; compinit" >> ~/.zshrc

# To load completions for each session, execute once:
$ yaml-crypt completion zsh > "${fpath[1]}/_yaml-crypt"

# You will need to start a new shell for this setup to take effect.

Fish:

$ yaml-crypt completion fish | source

# To load completions for each session, execute once:
$ yaml-crypt completion fish > ~/.config/fish/completions/yaml-crypt.fish
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionCmd)
}
