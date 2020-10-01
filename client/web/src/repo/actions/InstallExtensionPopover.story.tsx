import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../components/WebStory'
import { InstallExtensionPopover } from './InstallExtensionPopover'

const { add } = storiesOf('web/repo/actions', module).addDecorator(story => (
    <div className="container mt-3">{story()}</div>
))

add('InstallExtensionPopover (GitHub)', () => {
    function noopCallback() {
        // noop
    }

    return (
        <WebStory>
            {() => (
                <InstallExtensionPopover
                    url=""
                    serviceType="github"
                    onClose={noopCallback}
                    onRejection={noopCallback}
                    onClickInstall={noopCallback}
                    targetID="noop"
                />
            )}
        </WebStory>
    )
})
