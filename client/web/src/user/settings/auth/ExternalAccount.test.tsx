import { render } from '@testing-library/react'
import GithubIcon from 'mdi-react/GithubIcon'

import type { AuthProvider } from '../../../jscontext'

import { ExternalAccountConnectionDetails } from './ExternalAccount'

const mockAccount = {
    name: 'Github',
    icon: GithubIcon,
}

describe('ExternalAccountConnectionDetails', () => {
    test("renders correctly when display name isn't set", () => {
        const serviceTypes: AuthProvider['serviceType'][] = [
            'github',
            'gitlab',
            'bitbucketCloud',
            'http-header',
            'openidconnect',
            'sourcegraph-operator',
            'saml',
            'builtin',
            'gerrit',
            'azuredevops',
        ]

        for (const serviceType of serviceTypes) {
            const cmp = render(<ExternalAccountConnectionDetails account={mockAccount} serviceType={serviceType} />)
            expect(cmp.asFragment()).toMatchSnapshot()
        }
    })

    test('renders correctly when display name is set', () => {
        const serviceTypes: AuthProvider['serviceType'][] = [
            'github',
            'gitlab',
            'bitbucketCloud',
            'http-header',
            'openidconnect',
            'sourcegraph-operator',
            'saml',
            'builtin',
            'gerrit',
            'azuredevops',
        ]

        for (const serviceType of serviceTypes) {
            const cmp = render(
                <ExternalAccountConnectionDetails
                    account={{
                        ...mockAccount,
                        external: {
                            id: '123',
                            displayName: 'test@sourcegraph.com',
                            login: 'test',
                            url: 'https://example.com',
                        },
                    }}
                    serviceType={serviceType}
                />
            )
            expect(cmp.asFragment()).toMatchSnapshot()
        }
    })
})
