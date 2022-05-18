package main

import "fmt"

// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions

func ErrorMessage(title, message string) {
	fmt.Printf("::error title=%s::%s\n", title, message)
}

func NoticeMessage(title, message string) {
	fmt.Printf("::notice title=%s::%s\n", title, message)
}
