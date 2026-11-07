// clean-me is a safe, multi-stage system cleanup CLI. It orchestrates proven
// tools (docker, git, brew/apt/dnf, npm, pip) behind a confirm-first safety
// layer that always shows path · purpose · reclaimable size before deleting.
package main

import "github.com/ankitaniket/clean-me/cmd"

func main() {
	cmd.Execute()
}
