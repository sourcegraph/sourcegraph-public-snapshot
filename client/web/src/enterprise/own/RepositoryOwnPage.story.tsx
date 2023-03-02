import { Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'
import { MockedFeatureFlagsProvider } from '../../featureFlags/MockedFeatureFlagsProvider'
import { ExternalServiceKind, RepositoryFields } from '../../graphql-operations'

import { RepositoryOwnPage } from './RepositoryOwnPage'

const config: Meta = {
    title: 'web/enterprise/own/RepositoryOwnPage',
    parameters: {
        chromatic: { disableSnapshot: false },
    },
}

export default config

const repo: RepositoryFields = {
    id: '1',
    name: 'github.com/sourcegraph/sourcegraph',
    url: '/github.com/sourcegraph/sourcegraph',
    description: 'Code intelligence platform',
    externalRepository: {
        serviceID: '2',
        serviceType: 'github',
    },
    externalURLs: [
        {
            url: 'https://github.com/sourcegraph/sourcegraph',
            serviceKind: ExternalServiceKind.GITHUB,
        },
    ],
    viewerCanAdminister: false,
    defaultBranch: {
        displayName: 'main',
        abbrevName: 'main',
    },
}

export const EmptyNonAdmin: Story = () => (
    <WebStory>
        {({ useBreadcrumb }) => (
            <MockedFeatureFlagsProvider overrides={{ 'search-ownership': true }}>
                <RepositoryOwnPage repo={repo} authenticatedUser={{ siteAdmin: false }} useBreadcrumb={useBreadcrumb} />
            </MockedFeatureFlagsProvider>
        )}
    </WebStory>
)
EmptyNonAdmin.storyName = 'Empty (non-admin)'

export const EmptyAdmin: Story = () => (
    <WebStory>
        {({ useBreadcrumb }) => (
            <MockedFeatureFlagsProvider overrides={{ 'search-ownership': true }}>
                <RepositoryOwnPage repo={repo} authenticatedUser={{ siteAdmin: true }} useBreadcrumb={useBreadcrumb} />
            </MockedFeatureFlagsProvider>
        )}
    </WebStory>
)
EmptyAdmin.storyName = 'Empty (admin)'
