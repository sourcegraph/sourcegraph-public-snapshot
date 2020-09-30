import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../../components/WebStory'
import { CodeHostExtensionPopover } from './GoToCodeHostAction'

const { add } = storiesOf('web/repo/actions', module).addDecorator(story => (
    <div className="container mt-3">{story()}</div>
))

add('CodeHostExtensionPopover (GitHub)', () => {
    function noopCallback() {
        // noop
    }

    return (
        <WebStory>
            {() => (
                <CodeHostExtensionPopover
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
