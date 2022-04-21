import { select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { ExternalServiceKind } from '../../../graphql-operations'

import { AddCredentialModal } from './AddCredentialModal'
import { CREATE_BATCH_CHANGES_CREDENTIAL } from './backend'

const { add } = storiesOf('web/batches/settings/AddCredentialModal', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    })

add('Requires SSH - step 1', () => (
    <WebStory>
        {props => (
            <MockedTestProvider
                link={
                    new WildcardMockLink([
                        {
                            request: {
                                query: getDocumentNode(CREATE_BATCH_CHANGES_CREDENTIAL),
                                variables: MATCH_ANY_PARAMETERS,
                            },
                            result: {
                                data: {
                                    createBatchChangesCredential: {
                                        id: '123',
                                        isSiteCredential: false,
                                        sshPublicKey:
                                            'ssh-rsa randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
                                    },
                                },
                            },
                            nMatches: Number.POSITIVE_INFINITY,
                        },
                    ])
                }
            >
                <AddCredentialModal
                    {...props}
                    userID="user-id-1"
                    externalServiceKind={select(
                        'External service kind',
                        Object.values(ExternalServiceKind),
                        ExternalServiceKind.GITHUB
                    )}
                    externalServiceURL="https://github.com/"
                    requiresSSH={true}
                    requiresUsername={false}
                    afterCreate={noop}
                    onCancel={noop}
                />
            </MockedTestProvider>
        )}
    </WebStory>
))
add('Requires SSH - step 2', () => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={select(
                    'External service kind',
                    Object.values(ExternalServiceKind),
                    ExternalServiceKind.GITHUB
                )}
                externalServiceURL="https://github.com/"
                requiresSSH={true}
                requiresUsername={false}
                afterCreate={noop}
                onCancel={noop}
                initialStep="get-ssh-key"
            />
        )}
    </WebStory>
))

add('GitHub', () => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={ExternalServiceKind.GITHUB}
                externalServiceURL="https://github.com/"
                requiresSSH={false}
                requiresUsername={false}
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
))

add('GitLab', () => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={ExternalServiceKind.GITLAB}
                externalServiceURL="https://gitlab.com/"
                requiresSSH={false}
                requiresUsername={false}
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
))

add('Bitbucket Server', () => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={ExternalServiceKind.BITBUCKETSERVER}
                externalServiceURL="https://bitbucket.sgdev.org/"
                requiresSSH={false}
                requiresUsername={false}
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
))

add('Bitbucket Cloud', () => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={ExternalServiceKind.BITBUCKETCLOUD}
                externalServiceURL="https://bitbucket.org/"
                requiresSSH={false}
                requiresUsername={true}
                afterCreate={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
))
