import { DecoratorFn, Meta, Story } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'
import { BatchChangesCredentialFields, ExternalServiceKind } from '../../../graphql-operations'

import { ViewCredentialModal } from './ViewCredentialModal'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/ViewCredentialModal',
    decorators: [decorator],
}

export default config

const credential: BatchChangesCredentialFields = {
    id: '123',
    isSiteCredential: false,
    sshPublicKey:
        'ssh-rsa randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
}

export const ViewSupportedSignedCommits: Story = () => (
    <WebStory>
        {props => (
            <ViewCredentialModal
                {...props}
                codeHost={{
                    credential,
                    externalServiceKind: ExternalServiceKind.GITHUB,
                    externalServiceURL: 'https://github.com/',
                    requiresSSH: true,
                    requiresUsername: false,
                }}
                credential={credential}
                supportsSignedCommits={true}
                onClose={noop}
            />
        )}
    </WebStory>
)

export const ViewUnsupportedSignedCommits: Story = () => (
    <WebStory>
        {props => (
            <ViewCredentialModal
                {...props}
                codeHost={{
                    credential,
                    externalServiceKind: ExternalServiceKind.BITBUCKETSERVER,
                    externalServiceURL: 'https://bitbucket.sgdev.org/',
                    requiresSSH: true,
                    requiresUsername: false,
                }}
                credential={credential}
                supportsSignedCommits={false}
                onClose={noop}
            />
        )}
    </WebStory>
)

ViewSupportedSignedCommits.parameters = {
    chromatic: {
        // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
        delay: 2000,
    },
}
