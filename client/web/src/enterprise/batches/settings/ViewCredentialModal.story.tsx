import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'
import { BatchChangesCredentialFields, ExternalServiceKind } from '../../../graphql-operations'

import { ViewCredentialModal } from './ViewCredentialModal'

const { add } = storiesOf('web/batches/settings/ViewCredentialModal', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    })

const credential: BatchChangesCredentialFields = {
    id: '123',
    isSiteCredential: false,
    sshPublicKey:
        'ssh-rsa randorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorandorando',
}

add('View', () => (
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
                onClose={noop}
            />
        )}
    </WebStory>
))
