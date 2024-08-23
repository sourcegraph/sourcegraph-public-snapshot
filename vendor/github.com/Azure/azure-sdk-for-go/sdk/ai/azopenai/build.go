//go:build go1.18
// +build go1.18

// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

//go:generate pwsh ./testdata/genopenapi.ps1
//go:generate autorest  ./autorest.md
//go:generate go mod tidy
//go:generate goimports -w .

package azopenai
