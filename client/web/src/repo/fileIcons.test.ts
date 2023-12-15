import { describe, expect, it } from 'vitest'

import { FileExtension, containsTest, getFileInfo } from './fileIcons'

describe('getFileInfo', () => {
    const tests: {
        name: string
        file: string
        isDirectory: boolean
        expectedExtension: FileExtension
        expectedIsTest: boolean
    }[] = [
        {
            name: 'works with simple file name',
            file: 'my-file.js',
            isDirectory: false,
            expectedExtension: 'js' as FileExtension,
            expectedIsTest: false,
        },
        {
            name: 'works with complex file name',
            file: 'my-file.module.scss',
            isDirectory: false,
            expectedExtension: 'scss' as FileExtension,
            expectedIsTest: false,
        },
        {
            name: 'returns isTest as true if file name contains test',
            file: 'my-file.test.tsx',
            isDirectory: false,
            expectedExtension: 'tsx' as FileExtension,
            expectedIsTest: true,
        },
        {
            name: 'returns isTest as true if file name contains test',
            file: '.eslintrc',
            isDirectory: false,
            expectedExtension: 'default' as FileExtension,
            expectedIsTest: false,
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            const fileInfo = getFileInfo(t.file, t.isDirectory)
            expect(fileInfo.extension).toBe(t.expectedExtension)
            expect(fileInfo.isTest).toBe(t.expectedIsTest)
        })
    }
})

describe('containsTest', () => {
    const tests: {
        name: string
        file: string
        expected: boolean
    }[] = [
        {
            name: 'returns true if "test_" exists in file name',
            file: 'test_myfile.go',
            expected: true,
        },
        {
            name: 'returns true if "_test" exists in file name',
            file: 'myfile_test.go',
            expected: true,
        },
        {
            name: 'returns true if "_spec" exists in file name',
            file: 'myfile_spec.go',
            expected: true,
        },
        {
            name: 'returns true if "spec_" exists in file name',
            file: 'spec_myfile.go',
            expected: true,
        },
        {
            name: 'works with sub-extensions',
            file: 'myreactcomponent.test.tsx',
            expected: true,
        },
        {
            name: 'returns false if not a test file',
            file: 'mytestcomponent.java',
            expected: false,
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            expect(containsTest(t.file)).toBe(t.expected)
        })
    }
})
