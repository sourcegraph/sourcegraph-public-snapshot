import { describe, expect, it } from 'vitest'

import { isProbablyTestFile } from './icon-utils'

describe('isProbablyTestFile', () => {
    const tests: {
        file: string
        expected: boolean
    }[] = [
        {
            file: 'test_myfile.go',
            expected: true,
        },
        {
            file: 'myfile_test.go',
            expected: true,
        },
        {
            file: 'myfile_spec.go',
            expected: true,
        },
        {
            file: 'spec_myfile.go',
            expected: true,
        },
        {
            file: 'myreactcomponent.test.tsx',
            expected: true,
        },
        {
            file: 'mytestcomponent.java',
            expected: false,
        },
    ]

    for (const t of tests) {
        it(t.file, () => {
            expect(isProbablyTestFile(t.file)).toBe(t.expected)
        })
    }
})
