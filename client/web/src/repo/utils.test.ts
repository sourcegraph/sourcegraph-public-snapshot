import { describe, expect, it } from 'vitest'

import { FileExtension } from './constants'
import { contains, getExtension, getInitialSearchTerm } from './utils'

describe('contains', () => {
    const tests: {
        name: string
        collection: string[]
        target: string
        expected: boolean
    }[] = [
        {
            name: 'returns true if item exists in array',
            collection: ['bob', 'bill', 'sue', 'quinn', 'beyang'],
            target: 'sue',
            expected: true,
        },
        {
            name: 'returns false if item does not exist in array',
            collection: ['bob', 'bill', 'sue', 'quinn', 'beyang'],
            target: 'Taylor',
            expected: false,
        },
        {
            name: 'works on the first item',
            collection: ['bob', 'bill', 'sue', 'quinn', 'beyang'],
            target: 'bob',
            expected: true,
        },
        {
            name: 'works on the last item',
            collection: ['bob', 'bill', 'sue', 'quinn', 'beyang'],
            target: 'beyang',
            expected: true,
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            expect(contains(t.collection, t.target)).toBe(t.expected)
        })
    }
})

describe('getExtension', () => {
    const tests: {
        name: string
        file: string
        expectedExtension: FileExtension
        expectedIsTest: boolean
    }[] = [
        {
            name: 'works with simple file name',
            file: 'my-file.js',
            expectedExtension: 'js' as FileExtension,
            expectedIsTest: false,
        },
        {
            name: 'works with complex file name',
            file: 'my-file.module.scss',
            expectedExtension: 'scss' as FileExtension,
            expectedIsTest: false,
        },
        {
            name: 'returns isTest as true if file name contains test',
            file: 'my-file.test.tsx',
            expectedExtension: 'tsx' as FileExtension,
            expectedIsTest: true,
        },
        {
            name: "go.mod file returns 'go'",
            file: 'go.mod',
            expectedExtension: 'go' as FileExtension,
            expectedIsTest: false,
        },
        {
            name: "go.sum file returns 'go'",
            file: 'go.sum',
            expectedExtension: 'go' as FileExtension,
            expectedIsTest: false,
        },
    ]

    for (const t of tests) {
        it(t.name, () => {
            let e = getExtension(t.file)
            expect(e.extension).toBe(t.expectedExtension)
            expect(e.isTest).toBe(t.expectedIsTest)
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
