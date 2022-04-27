import { renderWithBrandedContext } from '../testing'

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
