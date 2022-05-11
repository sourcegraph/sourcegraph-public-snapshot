import { MockedResponse } from '@apollo/client/testing'
import { storiesOf } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import {
    BatchChangesCodeHostFields,
    ExternalServiceKind,
    GlobalBatchChangesCodeHostsResult,
} from '../../../graphql-operations'

import { GLOBAL_CODE_HOSTS } from './backend'
import { BatchChangesSiteConfigSettingsArea } from './BatchChangesSiteConfigSettingsArea'

const { add } = storiesOf('web/batches/settings/BatchChangesSiteConfigSettingsArea', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const createMock = (...hosts: BatchChangesCodeHostFields[]): MockedResponse<GlobalBatchChangesCodeHostsResult>[] => [
    {
        request: {
            query: getDocumentNode(GLOBAL_CODE_HOSTS),
            variables: {
                after: null,
                first: 15,
            },
        },
        result: {
            data: {
                batchChangesCodeHosts: {
                    totalCount: hosts.length,
                    pageInfo: { endCursor: null, hasNextPage: false },
                    nodes: hosts,
                },
            },
        },
    },
]

add('Overview', () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={createMock(
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
                        credential: null,
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
                )}
            >
                <BatchChangesSiteConfigSettingsArea {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('Config added', () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={createMock(
                    {
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
                    },
                    {
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
                    },
                    {
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
                    },
                    {
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
                    }
                )}
            >
                <BatchChangesSiteConfigSettingsArea {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
))
