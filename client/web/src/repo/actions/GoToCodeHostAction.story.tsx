import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React, { useEffect, useState } from 'react'
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
        {() => {
            const [open, setOpen] = useState(false)
            // The popover cannot be open on initial render
            // since the target element isn't in the DOM yet
            useEffect(() => {
                setTimeout(() => setOpen(true), 0)
            }, [])
            return (
                <>
                    <button
                        className="btn btn-outline-primary"
                        id="view-on-github"
                        onClick={() => setOpen(isOpen => !isOpen)}
                    >
                        View on GitHub
                    </button>
                    <InstallExtensionPopover
                        url=""
                        serviceType="github"
                        onClose={onClose}
                        onRejection={onRejection}
                        onClickInstall={onClickInstall}
                        targetID="view-on-github"
                        isOpen={open}
                        toggle={noop}
                    />
                </>
            )
        }}
    </WebStory>
))
