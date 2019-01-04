import React from 'react'
import renderer from 'react-test-renderer'
import { setLinkComponent } from './Link'
import { RepoLink } from './RepoLink'

describe('RepoLink', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => setLinkComponent(null as any)) // reset global env for other tests

    test('renders a link when "to" is set', () => {
        const component = renderer.create(<RepoLink repoName="my/repo" to="http://example.com" />)
        expect(component.toJSON()).toMatchSnapshot()
    })

    test('renders a fragment when "to" is null', () => {
        const component = renderer.create(<RepoLink repoName="my/repo" to={null} />)
        expect(component.toJSON()).toMatchSnapshot()
    })
})
