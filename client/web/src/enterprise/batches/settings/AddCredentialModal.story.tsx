import type { DecoratorFn, Story, Meta } from '@storybook/react'
import { noop } from 'lodash'
import { MATCH_ANY_PARAMETERS, WildcardMockLink } from 'wildcard-mock-link'

import { getDocumentNode } from '@sourcegraph/http-client'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { WebStory } from '../../../components/WebStory'
import { ExternalServiceKind } from '../../../graphql-operations'

import { AddCredentialModal } from './AddCredentialModal'
import { CREATE_BATCH_CHANGES_CREDENTIAL } from './backend'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/AddCredentialModal',
    decorators: [decorator],
    parameters: {
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    },
}

export default config

export const RequiresSSHstep1: Story = args => (
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
                    externalServiceKind={args.externalServiceKind}
                    externalServiceURL="https://github.com/"
                    requiresSSH={true}
                    requiresUsername={false}
                    afterCreate={noop}
                    onCancel={noop}
                />
            </MockedTestProvider>
        )}
    </WebStory>
)
RequiresSSHstep1.argTypes = {
    externalServiceKind: {
        name: 'External service kind',
        control: { type: 'select', options: Object.values(ExternalServiceKind) },
        defaultValue: ExternalServiceKind.GITHUB,
    },
}

RequiresSSHstep1.storyName = 'Requires SSH - step 1'

export const RequiresSSHstep2: Story = args => (
    <WebStory>
        {props => (
            <AddCredentialModal
                {...props}
                userID="user-id-1"
                externalServiceKind={args.externalServiceKind}
                externalServiceURL="https://github.com/"
                requiresSSH={true}
                requiresUsername={false}
                afterCreate={noop}
                onCancel={noop}
                initialStep="get-ssh-key"
            />
        )}
    </WebStory>
)
RequiresSSHstep2.argTypes = {
    externalServiceKind: {
        name: 'External service kind',
        control: { type: 'select', options: Object.values(ExternalServiceKind) },
        defaultValue: ExternalServiceKind.GITHUB,
    },
}

RequiresSSHstep2.storyName = 'Requires SSH - step 2'

export const GitHub: Story = () => (
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
)

GitHub.storyName = 'GitHub'

export const GitLab: Story = () => (
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
)

GitLab.storyName = 'GitLab'

export const BitbucketServer: Story = () => (
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
)

export const BitbucketCloud: Story = () => (
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
)
