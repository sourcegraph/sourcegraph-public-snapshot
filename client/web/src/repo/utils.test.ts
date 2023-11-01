import { describe, expect, it } from '@jest/globals'

import { getInitialRepoResults } from './utils'

describe('getInitialSearchTerm', () => {
    const tests: {
        name: string
        repo: string
        expected: String
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

    for (let i = 0; i < tests.length; i++) {
        let t = tests[i]
        it(t.name, () => {
            expect(getInitialRepoResults(t.repo)).toBe(t.expected)
        })
    }
})
