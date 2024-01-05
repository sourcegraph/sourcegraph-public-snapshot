import { render } from '@testing-library/react'
import { describe, expect, test } from 'vitest'

import { RepoLink, displayRepoName } from './RepoLink'

describe('RepoLink', () => {
    test('renders a link when "to" is set', () => {
        const component = render(<RepoLink repoName="my/repo" to="http://example.com" />)
        expect(component.asFragment()).toMatchSnapshot()
    })

    test('renders a fragment when "to" is null', () => {
        const component = render(<RepoLink repoName="my/repo" to={null} />)
        expect(component.asFragment()).toMatchSnapshot()
    })
})

describe('displayRepoName', () => {
    const testCases = [
        { originalRepoName: 'gerrit.sgdev.org/a/gabe/test', repoDisplayName: 'a/gabe/test' },
        { originalRepoName: 'github.com/sourcegraph/sourcegraph', repoDisplayName: 'sourcegraph/sourcegraph' },
        { originalRepoName: 'gerrit.sgdev.org/sourcegraph', repoDisplayName: 'sourcegraph' },
        { originalRepoName: 'sourcegraph', repoDisplayName: 'sourcegraph' },
        { originalRepoName: 'sourcegraph/sourcegraph', repoDisplayName: 'sourcegraph/sourcegraph' },
        { originalRepoName: 'sg.exe/sourcegraph', repoDisplayName: 'sourcegraph' },
        { originalRepoName: 'org.scala-sbt:collections_2.12', repoDisplayName: 'org.scala-sbt:collections_2.12' },
    ]

    for (const { originalRepoName, repoDisplayName } of testCases) {
        test(`displays ${repoDisplayName} for ${originalRepoName}`, () => {
            expect(displayRepoName(originalRepoName)).toEqual(repoDisplayName)
        })
    }
})
