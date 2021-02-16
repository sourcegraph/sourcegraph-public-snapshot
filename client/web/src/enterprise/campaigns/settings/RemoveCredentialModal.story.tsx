import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { RemoveCredentialModal } from './RemoveCredentialModal'

const { add } = storiesOf('web/campaigns/settings/RemoveCredentialModal', module)
    .addDecorator(story => <div className="p-3 container web-content">{story()}</div>)
    .addParameters({
        chromatic: {
            // Delay screenshot taking, so the modal has opened by the time the screenshot is taken.
            delay: 2000,
        },
    })

add('Confirmation', () => (
    <EnterpriseWebStory>
        {props => <RemoveCredentialModal {...props} credentialID="123" afterDelete={noop} onCancel={noop} />}
    </EnterpriseWebStory>
))
