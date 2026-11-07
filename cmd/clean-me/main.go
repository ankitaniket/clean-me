// clean-me is an interactive, safe, multi-stage system cleanup tool. Launched
// with no arguments it opens a terminal dashboard; its scan/clean/doctor
// subcommands provide a scriptable surface. It orchestrates proven tools
// (docker, git, brew/apt/dnf, npm, pip) behind a confirm-first safety layer.
package main

import "github.com/ankitaniket/clean-me/cmd"

func main() {
	cmd.Execute()
}
