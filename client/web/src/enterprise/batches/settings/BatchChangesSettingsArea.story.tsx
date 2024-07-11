import type { Decorator, Meta, StoryFn } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import {
    type BatchChangesCodeHostFields,
    type BatchChangesCredentialFields,
    ExternalServiceKind,
    type UserBatchChangesCodeHostsResult,
    type UserAreaUserFields,
} from '../../../graphql-operations'
import { BATCH_CHANGES_SITE_CONFIGURATION } from '../backend'
import { noRolloutWindowMockResult, rolloutWindowConfigMockResult } from '../mocks'

import { USER_CODE_HOSTS } from './backend'
import { BatchChangesSettingsArea } from './BatchChangesSettingsArea'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/BatchChangesSettingsArea',
    decorators: [decorator],
}

export default config

const codeHostsResult = (...hosts: BatchChangesCodeHostFields[]): UserBatchChangesCodeHostsResult => ({
    node: {
        __typename: 'User',
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
})

const sshCredential = (isSiteCredential: boolean): BatchChangesCredentialFields => ({
    id: '123',
    isSiteCredential,
    sshPublicKey:
        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
    gitHubApp: null,
})

export const Overview: StoryFn = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(USER_CODE_HOSTS),
                            variables: {
                                user: 'user-id-1',
                                after: null,
                                first: 15,
                            },
                        },
                        result: {
                            data: codeHostsResult(
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
                                    credential: sshCredential(true),
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
                        },
                    },
                    {
                        request: {
                            query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
                        },
                        result: noRolloutWindowMockResult,
                    },
                ]}
            >
                <BatchChangesSettingsArea {...props} user={{ id: 'user-id-1' } as UserAreaUserFields} />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const ConfigAdded: StoryFn = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(USER_CODE_HOSTS),
                            variables: {
                                user: 'user-id-2',
                                after: null,
                                first: 15,
                            },
                        },
                        result: {
                            data: codeHostsResult(
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: sshCredential(false),
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
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.GITLAB,
                                    externalServiceURL: 'https://gitlab.com/',
                                    requiresSSH: false,
                                    requiresUsername: false,
                                    supportsCommitSigning: false,
                                    commitSigningConfiguration: null,
                                },
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                    externalServiceURL: 'https://bitbucket.sgdev.org/',
                                    requiresSSH: true,
                                    requiresUsername: false,
                                    supportsCommitSigning: false,
                                    commitSigningConfiguration: null,
                                },
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.BITBUCKETCLOUD,
                                    externalServiceURL: 'https://bitbucket.org/',
                                    requiresSSH: false,
                                    requiresUsername: true,
                                    supportsCommitSigning: false,
                                    commitSigningConfiguration: null,
                                }
                            ),
                        },
                    },
                    {
                        request: {
                            query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
                        },
                        result: noRolloutWindowMockResult,
                    },
                ]}
            >
                <BatchChangesSettingsArea {...props} user={{ id: 'user-id-2' } as UserAreaUserFields} />
            </MockedTestProvider>
        )}
    </WebStory>
)

ConfigAdded.storyName = 'Config added'

export const RolloutWindowsConfigurationStory: StoryFn = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(USER_CODE_HOSTS),
                            variables: {
                                user: 'user-id-2',
                                after: null,
                                first: 15,
                            },
                        },
                        result: {
                            data: codeHostsResult(
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: sshCredential(false),
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
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.GITLAB,
                                    externalServiceURL: 'https://gitlab.com/',
                                    requiresSSH: false,
                                    requiresUsername: false,
                                    supportsCommitSigning: false,
                                    commitSigningConfiguration: null,
                                },
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                    externalServiceURL: 'https://bitbucket.sgdev.org/',
                                    requiresSSH: true,
                                    requiresUsername: false,
                                    supportsCommitSigning: false,
                                    commitSigningConfiguration: null,
                                },
                                {
                                    __typename: 'BatchChangesCodeHost',
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.BITBUCKETCLOUD,
                                    externalServiceURL: 'https://bitbucket.org/',
                                    requiresSSH: false,
                                    requiresUsername: true,
                                    supportsCommitSigning: false,
                                    commitSigningConfiguration: null,
                                }
                            ),
                        },
                    },
                    {
                        request: {
                            query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
                        },
                        result: rolloutWindowConfigMockResult,
                    },
                ]}
            >
                <BatchChangesSettingsArea {...props} user={{ id: 'user-id-2' } as UserAreaUserFields} />
            </MockedTestProvider>
        )}
    </WebStory>
)

RolloutWindowsConfigurationStory.storyName = 'Rollout Windows configured'
