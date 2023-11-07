import * as assert from 'assert'

import { describe, it } from 'vitest'

import { goSpec } from './go'
import { nilFilterContext, nilResult } from './spec.test'

const fileContent = `
package main

import "github.com/foo/foo/s"
import bar "github.com/foo/bar"
import (
	"github.com/foo/baz/s"
	quux "github.com/foo/honk/s/s"
)

func main() {
	// testing
}
`

describe('goSpec', () => {
    it('filters definitions', () => {
        const results = [
            { ...nilResult, repo: 'github.com/foo/foo', file: 's/a.go' },
            { ...nilResult, repo: 'github.com/foo/bar', file: 'b.go' },
            { ...nilResult, repo: 'github.com/foo/baz', file: 's/c.go' },
            { ...nilResult, repo: 'github.com/foo/honk', file: 's/s/d.go' },
            // same package
            { ...nilResult, repo: 'github.com/foo/test', file: 'x/y/w.go' },

            // incorrect repo
            { ...nilResult, repo: 'github.com/foo/quux', file: 'a.go' },
            // incorrect packages
            { ...nilResult, repo: 'github.com/foo/foo', file: 'a.go' },
            { ...nilResult, repo: 'github.com/foo/foo', file: 'r/a.go' },
        ]

        const filtered = goSpec.filterDefinitions?.(results, {
            ...nilFilterContext,
            repo: 'github.com/foo/test',
            filePath: 'x/y/z.go',
            fileContent,
        })

        assert.deepStrictEqual(filtered, [results[0], results[1], results[2], results[3], results[4]])
    })

    it('filters definitions from root package', () => {
        const results = [
            { ...nilResult, repo: 'github.com/foo/test', file: 'util.go' },

            // incorrect repo
            { ...nilResult, repo: 'github.com/ext/bar', file: 'a.go' },
        ]

        const filtered = goSpec.filterDefinitions?.(results, {
            ...nilFilterContext,
            repo: 'github.com/foo/test',
            filePath: 'main.go',
            fileContent,
        })

        assert.deepStrictEqual(filtered, [results[0]])
    })
})
