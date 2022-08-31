import * as assert from 'assert'

import { nilFilterContext, nilResult } from './spec.test'
import { typescriptSpec } from './typescript'

const fileContent = `
import { a, b, c } from "./bar"
const d = require('../../shared/baz')
`

describe('typescriptSpec', () => {
    it('filters definitions', () => {
        const results = [
            { ...nilResult, file: 'a/b/c/bar.ts' },
            { ...nilResult, file: 'a/b/c/bar.js' },
            { ...nilResult, file: 'a/shared/baz.ts' },
            { ...nilResult, file: 'a/shared/baz/index.ts' },

            // incorrect file
            { ...nilResult, file: 'a/b/c/baz.ts' },
            // incorrect paths
            { ...nilResult, file: 'x/y/z/bar.ts' },
            { ...nilResult, file: 'a/b/shared/baz.ts' },
        ]

        const filtered = typescriptSpec.filterDefinitions?.(results, {
            ...nilFilterContext,
            filePath: 'a/b/c/foo.ts',
            fileContent,
        })

        assert.deepStrictEqual(filtered, [results[0], results[1], results[2], results[3]])
    })
})
