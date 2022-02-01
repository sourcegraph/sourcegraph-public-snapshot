import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { ExternalServiceKind } from '@sourcegraph/shared/src/graphql-operations'

import { WebStory } from '../../../components/WebStory'

import { RemoveCredentialModal } from './RemoveCredentialModal'

const { add } = storiesOf('web/batches/settings/RemoveCredentialModal', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    })

const credential = {
    id: '123',
    isSiteCredential: false,
    sshPublicKey:
        'ssh-rsa randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
}

add('No ssh', () => (
    <WebStory>
        {props => (
            <RemoveCredentialModal
                {...props}
                codeHost={{
                    credential,
                    requiresSSH: false,
                    externalServiceKind: ExternalServiceKind.GITHUB,
                    externalServiceURL: 'https://github.com/',
                }}
                credential={credential}
                afterDelete={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
))

add('Requires ssh', () => (
    <WebStory>
        {props => (
            <RemoveCredentialModal
                {...props}
                codeHost={{
                    credential,
                    requiresSSH: true,
                    externalServiceKind: ExternalServiceKind.GITHUB,
                    externalServiceURL: 'https://github.com/',
                }}
                credential={credential}
                afterDelete={noop}
                onCancel={noop}
            />
        )}
    </WebStory>
))
