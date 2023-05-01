import { MockedResponse } from '@apollo/client/testing'
import { Meta, Story } from '@storybook/react'
import { subDays } from 'date-fns'

import { getDocumentNode } from '@sourcegraph/http-client'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { WebStory } from '../../components/WebStory'
import { createFlagMock } from '../../featureFlags/createFlagMock'
import {
    ExternalServiceKind,
    GetIngestedCodeownersResult,
    GetIngestedCodeownersVariables,
    RepositoryFields,
} from '../../graphql-operations'

import { GET_INGESTED_CODEOWNERS_QUERY } from './graphqlQueries'
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
    isFork: false,
    viewerCanAdminister: false,
    defaultBranch: {
        displayName: 'main',
        abbrevName: 'main',
    },
    metadata: [],
}

const empyResponse: MockedResponse<GetIngestedCodeownersResult, GetIngestedCodeownersVariables> = {
    request: {
        query: getDocumentNode(GET_INGESTED_CODEOWNERS_QUERY),
        variables: {
            repoID: repo.id,
        },
    },
    result: {
        data: { node: { ingestedCodeowners: null, __typename: 'Repository' } },
    },
}

const searchOwnershipFlagMock = createFlagMock('search-ownership', true)

export const EmptyNonAdmin: Story = () => (
    <WebStory mocks={[empyResponse, searchOwnershipFlagMock]}>
        {({ useBreadcrumb }) => (
            <RepositoryOwnPage
                repo={repo}
                authenticatedUser={{ siteAdmin: false }}
                useBreadcrumb={useBreadcrumb}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
)
EmptyNonAdmin.storyName = 'Empty (non-admin)'

export const EmptyAdmin: Story = () => (
    <WebStory mocks={[empyResponse, searchOwnershipFlagMock]}>
        {({ useBreadcrumb }) => (
            <RepositoryOwnPage
                repo={repo}
                authenticatedUser={{ siteAdmin: true }}
                useBreadcrumb={useBreadcrumb}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
)
EmptyAdmin.storyName = 'Empty (admin)'

const populatedResponse: MockedResponse<GetIngestedCodeownersResult, GetIngestedCodeownersVariables> = {
    request: {
        query: getDocumentNode(GET_INGESTED_CODEOWNERS_QUERY),
        variables: {
            repoID: repo.id,
        },
    },
    result: {
        data: {
            node: {
                ingestedCodeowners: {
                    contents:
                        '/.github/workflows/codenotify.yml @unknwon\n/.github/workflows/licenses-check.yml @bobheadxi\n/.github/workflows/licenses-update.yml @bobheadxi\n/.github/workflows/renovate-downstream.yml @bobheadxi\n/.github/workflows/renovate-downstream.json @bobheadxi\n',
                    updatedAt: subDays(new Date(), 1).toISOString(),
                    __typename: 'CodeownersIngestedFile',
                },
                __typename: 'Repository',
            },
        },
    },
}

export const PopulatedNonAdmin: Story = () => (
    <WebStory mocks={[populatedResponse, searchOwnershipFlagMock]}>
        {({ useBreadcrumb }) => (
            <RepositoryOwnPage
                repo={repo}
                authenticatedUser={{ siteAdmin: false }}
                useBreadcrumb={useBreadcrumb}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
)
PopulatedNonAdmin.storyName = 'Populated (non-admin)'

export const PopulatedAdmin: Story = () => (
    <WebStory mocks={[populatedResponse, searchOwnershipFlagMock]}>
        {({ useBreadcrumb }) => (
            <RepositoryOwnPage
                repo={repo}
                authenticatedUser={{ siteAdmin: true }}
                useBreadcrumb={useBreadcrumb}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )}
    </WebStory>
)
PopulatedAdmin.storyName = 'Populated (admin)'
