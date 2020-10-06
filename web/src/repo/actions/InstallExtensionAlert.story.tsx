import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../components/WebStory'
import { InstallExtensionAlert } from './InstallExtensionAlert'

const onAlertDismissed = action('onAlertDismissed')

const { add } = storiesOf('web/repo/actions', module).addDecorator(story => (
    <div className="container mt-3">{story()}</div>
))

add('InstallExtensionAlert (GitHub)', () => (
    <WebStory>
        {() => (
            <InstallExtensionAlert
                onAlertDismissed={onAlertDismissed}
                externalURLs={[
                    {
                        __typename: 'ExternalLink',
                        url: 'https://github.com/sourcegraph/sourcegraph',
                        serviceType: 'github',
                    },
                ]}
            />
        )}
    </WebStory>
))
