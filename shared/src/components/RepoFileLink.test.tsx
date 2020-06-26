import React from 'react'
import { RepoFileLink } from './RepoFileLink'
import { mount } from 'enzyme'

describe('RepoFileLink', () => {
    test('renders', () => {
        const component = mount(
            <RepoFileLink
                repoName="example.com/my/repo"
                repoURL="https://example.com"
                filePath="my/file"
                fileURL="https://example.com/file"
            />
        )
        expect(component.children()).toMatchSnapshot()
    })
})
