import { MockedResponse } from '@apollo/client/testing'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import {
    BatchChangesCodeHostFields,
    ExternalServiceKind,
    GlobalBatchChangesCodeHostsResult,
} from '../../../graphql-operations'
import { BATCH_CHANGES_SITE_CONFIGURATION } from '../backend'
import { rolloutWindowConfigMockResult } from '../mocks'

import { GLOBAL_CODE_HOSTS } from './backend'
import { BatchChangesSiteConfigSettingsArea } from './BatchChangesSiteConfigSettingsArea'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/BatchChangesSiteConfigSettingsArea',
    decorators: [decorator],
}

export default config

const ROLLOUT_WINDOWS_CONFIGURATION_MOCK = {
    request: {
        query: getDocumentNode(BATCH_CHANGES_SITE_CONFIGURATION),
    },
    result: rolloutWindowConfigMockResult,
}

const createMock = (...hosts: BatchChangesCodeHostFields[]): MockedResponse<GlobalBatchChangesCodeHostsResult> => ({
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
})

export const Overview: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    ROLLOUT_WINDOWS_CONFIGURATION_MOCK,
                    createMock(
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
                    ),
                ]}
            >
                <BatchChangesSiteConfigSettingsArea {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
)

export const ConfigAdded: Story = () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    ROLLOUT_WINDOWS_CONFIGURATION_MOCK,
                    createMock(
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
                    ),
                ]}
            >
                <BatchChangesSiteConfigSettingsArea {...props} />
            </MockedTestProvider>
        )}
    </WebStory>
)

ConfigAdded.storyName = 'Config added'
