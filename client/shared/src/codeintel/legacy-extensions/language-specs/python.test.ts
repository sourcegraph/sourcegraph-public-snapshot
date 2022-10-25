import * as assert from 'assert'

import { pythonSpec, relativeImportPath } from './python'
import { nilFilterContext, nilResult } from './spec.test'

const fileContent = `
import .bar
from .baz.bonk import honk
from ..bonk.quux import ronk
import a.b.c.quux
`

describe('pythonSpec', () => {
    it('filters definitions', () => {
        const results = [
            { ...nilResult, file: 'a/b/c/bar.py' },
            { ...nilResult, file: 'a/b/c/baz/bonk.py' },
            { ...nilResult, file: 'a/b/bonk/quux.py' },
            { ...nilResult, file: 'a/b/c/quux.py' },

            // incorrect file
            { ...nilResult, file: 'a/b/c/baz.py' },
            // incorrect paths
            { ...nilResult, file: 'a/b/quux.py' },
            { ...nilResult, file: 'x/y/z/quux.py' },
        ]

        const filtered = pythonSpec.filterDefinitions?.(results, {
            ...nilFilterContext,
            filePath: 'a/b/c/foo.py',
            fileContent,
        })

        assert.deepStrictEqual(filtered, [results[0], results[1], results[2], results[3]])
    })
})

describe('relativeImportPath', () => {
    it('converts relative import to path', () => {
        assert.strictEqual(relativeImportPath('a/b/c.py', 'foo'), undefined)
        assert.strictEqual(relativeImportPath('a/b/c.py', '.foo'), 'a/b/foo')
        assert.strictEqual(relativeImportPath('a/b/c.py', '.foo.bar'), 'a/b/foo/bar')
        assert.strictEqual(relativeImportPath('a/b/c.py', '..foo.bar'), 'a/foo/bar')
        assert.strictEqual(relativeImportPath('a/b/c.py', '....foo'), '../foo')
    })
})
