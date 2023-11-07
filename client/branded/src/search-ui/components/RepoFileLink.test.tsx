import { describe, expect, test } from '@jest/globals'

import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { RepoFileLink } from './RepoFileLink'

describe('RepoFileLink', () => {
    test('renders', () => {
        const component = renderWithBrandedContext(
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
