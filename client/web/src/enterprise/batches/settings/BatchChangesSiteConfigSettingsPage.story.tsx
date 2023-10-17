import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { MATCH_ANY_PARAMETERS, type WildcardMockedResponse, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { type BatchChangesCodeHostFields, ExternalServiceKind } from '../../../graphql-operations'
import { BATCH_CHANGES_SITE_CONFIGURATION } from '../backend'
import { rolloutWindowConfigMockResult } from '../mocks'

import { GLOBAL_CODE_HOSTS } from './backend'
import { BatchChangesSiteConfigSettingsPage } from './BatchChangesSiteConfigSettingsPage'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/BatchChangesSiteConfigSettingsPage',
    decorators: [decorator],
}

export default config

const ROLLOUT_WINDOWS_CONFIGURATION_MOCK = {
    request: {
        query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
    },
    result: rolloutWindowConfigMockResult,
    nMatches: Number.POSITIVE_INFINITY,
}

const createMock = (...hosts: BatchChangesCodeHostFields[]): WildcardMockedResponse => ({
    request: {
        query: getDocumentNode(GLOBAL_CODE_HOSTS),
        variables: MATCH_ANY_PARAMETERS,
    },
    result: {
        data: {
            batchChangesCodeHosts: {
                __typename: 'BatchChangesCodeHostConnection',
                totalCount: hosts.length,
                pageInfo: {
                    endCursor: null,
                    hasNextPage: false,
                    __typename: 'PageInfo',
                },
                nodes: hosts,
            },
        },
    },
    nMatches: Number.POSITIVE_INFINITY,
})

export const Overview: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink(
                        [
                            ROLLOUT_WINDOWS_CONFIGURATION_MOCK,
                            createMock(
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: null,
                                    externalServiceKind: ExternalServiceKind.GITHUB,
                                    externalServiceURL: 'https://github.com/',
                                    requiresSSH: false,
                                    requiresUsername: false,
                                    supportsCommitSigning: true,
                                    commitSigningConfiguration: {
                                        __typename: 'GitHubApp',
                                        id: '123',
                                        appID: 123,
                                        name: 'Sourcegraph Commit Signing',
                                        appURL: 'https://github.com/apps/sourcegraph-commit-signing',
                                        baseURL: 'https://github.com/',
                                        logo: 'https://github.com/identicons/app/app/commit-testing-local',
                                    },
                                },
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: null,
                                    externalServiceKind: ExternalServiceKind.GITHUB,
                                    externalServiceURL: 'https://github.mycompany.com/',
                                    requiresSSH: false,
                                    requiresUsername: false,
                                    supportsCommitSigning: true,
                                    commitSigningConfiguration: null,
                                },
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: null,
                                    externalServiceKind: ExternalServiceKind.GITLAB,
                                    externalServiceURL: 'https://gitlab.com/',
                                    requiresSSH: false,
                                    requiresUsername: false,
                                    supportsCommitSigning: false,
                                    commitSigningConfiguration: null,
                                },
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: null,
                                    externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                    externalServiceURL: 'https://bitbucket.sgdev.org/',
                                    requiresSSH: true,
                                    requiresUsername: false,
                                    supportsCommitSigning: false,
                                    commitSigningConfiguration: null,
                                },
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: null,
                                    externalServiceKind: ExternalServiceKind.BITBUCKETCLOUD,
                                    externalServiceURL: 'https://bitbucket.org/',
                                    requiresSSH: false,
                                    requiresUsername: true,
                                    supportsCommitSigning: false,
                                    commitSigningConfiguration: null,
                                }
                            ),
                        ],
                        { addTypename: true }
                    )
                }
            >
                <BatchChangesSiteConfigSettingsPage />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const NoItems: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider link={new WildcardMockLink([ROLLOUT_WINDOWS_CONFIGURATION_MOCK, createMock()])}>
                <BatchChangesSiteConfigSettingsPage />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const ConfigAdded: StoryFn = () => (
    <WebStory>
        {() => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        ROLLOUT_WINDOWS_CONFIGURATION_MOCK,
                        createMock(
                            {
                                __typename: 'BatchChangesCodeHost',
                                credential: {
                                    id: '123',
                                    isSiteCredential: true,
                                    sshPublicKey:
                                        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
                                },
                                externalServiceKind: ExternalServiceKind.GITHUB,
                                externalServiceURL: 'https://github.com/',
                                requiresSSH: false,
                                requiresUsername: false,
                                supportsCommitSigning: true,
                                commitSigningConfiguration: {
                                    __typename: 'GitHubApp',
                                    id: '123',
                                    appID: 123,
                                    name: 'Sourcegraph Commit Signing',
                                    appURL: 'https://github.com/apps/sourcegraph-commit-signing',
                                    baseURL: 'https://github.com/',
                                    logo: 'https://github.com/identicons/app/app/commit-testing-local',
                                },
                            },
                            {
                                __typename: 'BatchChangesCodeHost',
                                credential: {
                                    id: '123',
                                    isSiteCredential: true,
                                    sshPublicKey:
                                        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
                                },
                                externalServiceKind: ExternalServiceKind.GITLAB,
                                externalServiceURL: 'https://gitlab.com/',
                                requiresSSH: false,
                                requiresUsername: false,
                                supportsCommitSigning: false,
                                commitSigningConfiguration: null,
                            },
                            {
                                __typename: 'BatchChangesCodeHost',
                                credential: {
                                    id: '123',
                                    isSiteCredential: true,
                                    sshPublicKey:
                                        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
                                },
                                externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                externalServiceURL: 'https://bitbucket.sgdev.org/',
                                requiresSSH: true,
                                requiresUsername: false,
                                supportsCommitSigning: false,
                                commitSigningConfiguration: null,
                            },
                            {
                                __typename: 'BatchChangesCodeHost',
                                credential: {
                                    id: '123',
                                    isSiteCredential: true,
                                    sshPublicKey:
                                        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
                                },
                                externalServiceKind: ExternalServiceKind.BITBUCKETCLOUD,
                                externalServiceURL: 'https://bitbucket.org/',
                                requiresSSH: false,
                                requiresUsername: true,
                                supportsCommitSigning: false,
                                commitSigningConfiguration: null,
                            }
                        ),
                    ])
                }
            >
                <BatchChangesSiteConfigSettingsPage />
            </MockedTestProvider>
        )}
    </WebStory>
)

ConfigAdded.storyName = 'Config added'
