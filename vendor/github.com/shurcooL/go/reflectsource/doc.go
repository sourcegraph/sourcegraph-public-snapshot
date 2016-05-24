// Package sourcereflect implements run-time source reflection, allowing a program to
// look up string representation of objects from the underlying .go source files.
//
// Specifically, it implements ability to get name of caller funcs and their parameters.
// It also implements functionality to get a string containing source code of provided func.
//
// In order to succeed, it expects the program's source code to be available in normal location.
// It's intended to be used for development purposes, or for experimental programs.
package reflectsource
