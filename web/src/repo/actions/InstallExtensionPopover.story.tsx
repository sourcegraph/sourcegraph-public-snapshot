import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../components/WebStory'
import { InstallExtensionPopover } from './InstallExtensionPopover'

const onClose = action('onClose')
const onRejection = action('onRejection')
const onClickInstall = action('onClickInstall')

const { add } = storiesOf('web/repo/actions', module).addDecorator(story => (
    <div className="container mt-3">{story()}</div>
))

add('InstallExtensionPopover (GitHub)', () => (
    <WebStory>
        {() => (
            <InstallExtensionPopover
                url=""
                serviceType="github"
                onClose={onClose}
                onRejection={onRejection}
                onClickInstall={onClickInstall}
                targetID="noop"
            />
        )}
    </WebStory>
))
