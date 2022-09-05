import { render } from '@testing-library/react'

import { RepoLink } from './RepoLink'

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
