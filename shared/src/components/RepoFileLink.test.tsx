import React from 'react'
import renderer from 'react-test-renderer'
import { setLinkComponent } from './Link'
import { RepoFileLink } from './RepoFileLink'

describe('RepoFileLink', () => {
    setLinkComponent((props: any) => <a {...props} />)
    afterAll(() => setLinkComponent(null as any)) // reset global env for other tests

    test('renders', () => {
        const component = renderer.create(
            <RepoFileLink
                repoName="example.com/my/repo"
                repoURL="https://example.com"
                filePath="my/file"
                fileURL="https://example.com/file"
            />
        )
        expect(component.toJSON()).toMatchSnapshot()
    })
})
