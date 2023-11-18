import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'
import { type BatchChangesCredentialFields, ExternalServiceKind } from '../../../graphql-operations'

import { ViewCredentialModal } from './ViewCredentialModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

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

export const View: StoryFn = () => (
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
                    supportsCommitSigning: false,
                    commitSigningConfiguration: null,
                }}
                credential={credential}
                onClose={noop}
            />
        )}
    </WebStory>
)

View.parameters = {
    chromatic: {
        // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
        delay: 2000,
    },
}
