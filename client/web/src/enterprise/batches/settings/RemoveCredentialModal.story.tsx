import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../components/WebStory'

import { RemoveCredentialModal } from './RemoveCredentialModal'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/settings/RemoveCredentialModal',
    decorators: [decorator],
    parameters: {
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    },
}

export default config

const credential = {
    id: '123',
    isSiteCredential: false,
    sshPublicKey:
        'ssh-rsa randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
}

export const NoSsh: StoryFn = () => (
    <WebStory>
        {props => (
            <RemoveCredentialModal
                {...props}
                codeHost={{
                    credential,
                    requiresSSH: false,
                    requiresUsername: false,
                    externalServiceKind: ExternalServiceKind.GITHUB,
                    externalServiceURL: 'https://github.com/',
                    supportsCommitSigning: false,
                    commitSigningConfiguration: null,
                }}
                credential={credential}
                afterDelete={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
)

NoSsh.storyName = 'No ssh'

export const RequiresSsh: StoryFn = () => (
    <WebStory>
        {props => (
            <RemoveCredentialModal
                {...props}
                codeHost={{
                    credential,
                    requiresSSH: true,
                    requiresUsername: false,
                    externalServiceKind: ExternalServiceKind.GITHUB,
                    externalServiceURL: 'https://github.com/',
                    supportsCommitSigning: false,
                    commitSigningConfiguration: null,
                }}
                credential={credential}
                afterDelete={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
)

RequiresSsh.storyName = 'Requires ssh'
