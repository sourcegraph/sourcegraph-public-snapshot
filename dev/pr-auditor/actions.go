package main

import "fmt"

// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions

func setError(title, message string) {
	fmt.Printf("::error title=%s::%s\n", title, message)
}

func setNotice(title, message string) {
	fmt.Printf("::info title=%s::%s\n", title, message)
}
