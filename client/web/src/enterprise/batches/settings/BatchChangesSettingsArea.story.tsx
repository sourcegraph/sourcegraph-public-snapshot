import { storiesOf } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import {
    BatchChangesCodeHostFields,
    BatchChangesCredentialFields,
    ExternalServiceKind,
    UserBatchChangesCodeHostsResult,
} from '../../../graphql-operations'

import { USER_CODE_HOSTS } from './backend'
import { BatchChangesSettingsArea } from './BatchChangesSettingsArea'

const { add } = storiesOf('web/batches/settings/BatchChangesSettingsArea', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const codeHostsResult = (...hosts: BatchChangesCodeHostFields[]): UserBatchChangesCodeHostsResult => ({
    node: {
        __typename: 'User',
        batchChangesCodeHosts: {
            totalCount: hosts.length,
            pageInfo: { endCursor: null, hasNextPage: false },
            nodes: hosts,
        },
    },
})

const sshCredential = (isSiteCredential: boolean): BatchChangesCredentialFields => ({
    id: '123',
    isSiteCredential,
    sshPublicKey:
        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
})

add('Overview', () => (
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
                                    credential: null,
                                    externalServiceKind: ExternalServiceKind.GITHUB,
                                    externalServiceURL: 'https://github.com/',
                                    requiresSSH: false,
                                    requiresUsername: false,
                                },
                                {
                                    credential: null,
                                    externalServiceKind: ExternalServiceKind.GITLAB,
                                    externalServiceURL: 'https://gitlab.com/',
                                    requiresSSH: false,
                                    requiresUsername: false,
                                },
                                {
                                    credential: sshCredential(true),
                                    externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                    externalServiceURL: 'https://bitbucket.sgdev.org/',
                                    requiresSSH: true,
                                    requiresUsername: false,
                                },
                                {
                                    credential: null,
                                    externalServiceKind: ExternalServiceKind.BITBUCKETCLOUD,
                                    externalServiceURL: 'https://bitbucket.org/',
                                    requiresSSH: false,
                                    requiresUsername: true,
                                }
                            ),
                        },
                    },
                ]}
            >
                <BatchChangesSettingsArea {...props} user={{ id: 'user-id-1' }} />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('Config added', () => (
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
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.GITHUB,
                                    externalServiceURL: 'https://github.com/',
                                    requiresSSH: false,
                                    requiresUsername: false,
                                },
                                {
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.GITLAB,
                                    externalServiceURL: 'https://gitlab.com/',
                                    requiresSSH: false,
                                    requiresUsername: false,
                                },
                                {
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                                    externalServiceURL: 'https://bitbucket.sgdev.org/',
                                    requiresSSH: true,
                                    requiresUsername: false,
                                },
                                {
                                    credential: sshCredential(false),
                                    externalServiceKind: ExternalServiceKind.BITBUCKETCLOUD,
                                    externalServiceURL: 'https://bitbucket.org/',
                                    requiresSSH: false,
                                    requiresUsername: true,
                                }
                            ),
                        },
                    },
                ]}
            >
                <BatchChangesSettingsArea {...props} user={{ id: 'user-id-2' }} />
            </MockedTestProvider>
        )}
    </WebStory>
))
