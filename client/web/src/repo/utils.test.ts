import { describe, expect, it } from 'vitest'

import { FileExtension, containsTest, getFileInfo } from './fileIcons'
import { getInitialSearchTerm } from './utils'

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

describe('getInitialSearchTerm', () => {
    const tests: {
        name: string
        repo: string
        expected: string
    }[] = [
        {
            name: 'works with a github repo url',
            repo: 'github.com/sourcegraph/sourcegraph',
            expected: 'sourcegraph',
        },
        {
            name: 'works with a gitlab repo url',
            repo: 'gitlab.com/SourcegraphCody/jsonrpc2',
            expected: 'jsonrpc2',
        },
        {
            name: 'works with a perforce depot url',
            repo: 'public.perforce.com/sourcegraph/myp4depot',
            expected: 'myp4depot',
        },
        {
            name: 'works with a bitbucket repo name',
            repo: 'bitbucket.org/username/projectname/mybitbucketrepo',
            expected: 'mybitbucketrepo',
        },
        {
            name: 'works with a gerrit repo name',
            repo: 'mygerritserver.com/c/mygerritrepo',
            expected: 'mygerritrepo',
        },
        {
            name: 'works with an Azure DevOps repo name',
            repo: 'https://dev.azure.com/myADOorgname/myADOproject/_git/myADOrepo',
            expected: 'myADOrepo',
        },
        {
            name: 'works with a Plastic SCM repo name',
            repo: 'https://cloud.plasticscm.com/my-plastic-repo',
            expected: 'my-plastic-repo',
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            expect(getInitialSearchTerm(t.repo)).toBe(t.expected)
        })
    }
})
