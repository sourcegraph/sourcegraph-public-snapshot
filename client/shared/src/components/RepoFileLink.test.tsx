import { render } from '@testing-library/react'
import React from 'react'

import { RepoFileLink } from './RepoFileLink'

describe('RepoFileLink', () => {
    test('renders', () => {
        const component = render(
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
