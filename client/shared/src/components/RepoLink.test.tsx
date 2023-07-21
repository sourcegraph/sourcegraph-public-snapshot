import { render } from '@testing-library/react'

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
    test.each([
        ['gerrit.sgdev.org/a/gabe/test', 'a/gabe/test'],
        ['github.com/sourcegraph/sourcegraph', 'sourcegraph/sourcegraph'],
        ['gerrit.sgdev.org/sourcegraph', 'sourcegraph'],
        ['sourcegraph', 'sourcegraph'],
        ['sourcegraph/sourcegraph', 'sourcegraph/sourcegraph'],
        ['sg.exe/sourcegraph', 'sourcegraph'],
    ])('should return repo name correctly', (repoName: string, result: string) => {
        const name = displayRepoName(repoName)
        expect(name).toEqual(result)
    })
})
