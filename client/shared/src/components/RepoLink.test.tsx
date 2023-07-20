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
    test('removes code host from repo name with >= 3 slashes', () => {
        const name = displayRepoName('gerrit.sgdev.org/a/gabe/test')
        expect(name).toEqual('a/gabe/test')
    })

    test('removes code host from repo name with 2 slashes', () => {
        const name = displayRepoName('github.com/sourcegraph/sourcegraph')
        expect(name).toEqual('sourcegraph/sourcegraph')
    })

    test('removes code host from repo name with one slash', () => {
        const name = displayRepoName('gerrit.sgdev.org/sourcegraph')
        expect(name).toEqual('sourcegraph')
    })

    test('returns repo name when code host information is unavailable', () => {
        const name = displayRepoName('sourcegraph')
        expect(name).toEqual('sourcegraph')
    })
})
