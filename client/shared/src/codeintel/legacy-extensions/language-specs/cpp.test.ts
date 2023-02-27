import * as assert from 'assert'

import { cppSpec } from './cpp'
import { nilFilterContext, nilResult } from './spec.test'

const fileContent = `
@import a.b.c;
#import "d/e/f.h"
#include "g/h/i"
`

describe('cppSpec', () => {
    it('filters definitions', () => {
        const results = [
            { ...nilResult, file: 'a/b/c' },
            { ...nilResult, file: 'foo/bar/a/b/c.ext' },
            { ...nilResult, file: 'd/e/f.cpp' },
            { ...nilResult, file: 'g/h/i.h' },
            // no path segments in common
            { ...nilResult, file: 'x/y/z.cpp' },
            // no proper suffix imported
            { ...nilResult, file: 'e/f.cpp' },
        ]

        const filtered = cppSpec.filterDefinitions?.(results, {
            ...nilFilterContext,
            fileContent,
        })

        assert.deepStrictEqual(filtered, [results[0], results[1], results[2], results[3]])
    })
})
