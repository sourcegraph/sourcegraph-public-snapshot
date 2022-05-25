import { storiesOf } from '@storybook/react'
import { GraphQLError } from 'graphql'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import {
    BatchChangesCredentialFields,
    CheckBatchChangesCredentialResult,
    ExternalServiceKind,
} from '../../../graphql-operations'

import { CHECK_BATCH_CHANGES_CREDENTIAL } from './backend'
import { CodeHostConnectionNode } from './CodeHostConnectionNode'

const { add } = storiesOf('web/batches/settings/CodeHostConnectionNode', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const checkCredResult = (): CheckBatchChangesCredentialResult => ({
    checkBatchChangesCredential: {
        alwaysNil: null,
    },
})

const checkCredErrors = (): GraphQLError[] => [new GraphQLError('Credential is not valid')]

const sshCredential = (isSiteCredential: boolean): BatchChangesCredentialFields => ({
    id: '123',
    isSiteCredential,
    sshPublicKey:
        'rsa-ssh randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
})

add('Not Checked', () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(CHECK_BATCH_CHANGES_CREDENTIAL),
                            variables: {
                                id: '123',
                            },
                        },
                        result: {
                            data: checkCredResult(),
                        },
                    },
                ]}
            >
                <CodeHostConnectionNode
                    {...props}
                    node={{
                        credential: sshCredential(false),
                        externalServiceKind: ExternalServiceKind.GITHUB,
                        externalServiceURL: 'https://github.com/',
                        requiresSSH: false,
                        requiresUsername: false,
                    }}
                    refetchAll={() => {}}
                    userID="123"
                />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('Checked and Valid', () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(CHECK_BATCH_CHANGES_CREDENTIAL),
                            variables: {
                                id: '123',
                            },
                        },
                        result: {
                            data: checkCredResult(),
                        },
                    },
                ]}
            >
                <CodeHostConnectionNode
                    {...props}
                    node={{
                        credential: sshCredential(false),
                        externalServiceKind: ExternalServiceKind.GITHUB,
                        externalServiceURL: 'https://github.com/',
                        requiresSSH: false,
                        requiresUsername: false,
                    }}
                    refetchAll={() => {}}
                    userID="123"
                />
            </MockedTestProvider>
        )}
    </WebStory>
))

add('Checked and Invalid', () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                mocks={[
                    {
                        request: {
                            query: getDocumentNode(CHECK_BATCH_CHANGES_CREDENTIAL),
                            variables: {
                                id: '123',
                            },
                        },
                        result: {
                            errors: checkCredErrors(),
                        },
                    },
                ]}
            >
                <CodeHostConnectionNode
                    {...props}
                    node={{
                        credential: sshCredential(false),
                        externalServiceKind: ExternalServiceKind.GITHUB,
                        externalServiceURL: 'https://github.com/',
                        requiresSSH: false,
                        requiresUsername: false,
                    }}
                    refetchAll={() => {}}
                    userID="123"
                />
            </MockedTestProvider>
        )}
    </WebStory>
))
