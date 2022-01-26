import React from 'react'

import { renderWithRouter } from '../testing/render-with-router'

import { RepoFileLink } from './RepoFileLink'

describe('RepoFileLink', () => {
    test('renders', () => {
        const component = renderWithRouter(
            <RepoFileLink
                repoName="example.com/my/repo"
                repoURL="https://example.com"
                filePath="my/file"
                fileURL="https://example.com/file"
            />
        )
        expect(component.asFragment()).toMatchSnapshot()
    })
})
