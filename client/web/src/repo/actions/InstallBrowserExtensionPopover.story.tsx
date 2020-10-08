import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React, { useEffect, useState } from 'react'
import { WebStory } from '../../components/WebStory'
import { InstallBrowserExtensionPopover } from './InstallBrowserExtensionPopover'

const onClose = action('onClose')
const onRejection = action('onRejection')
const onClickInstall = action('onClickInstall')

const { add } = storiesOf('web/repo/actions/InstallBrowserExtensionPopover', module).addDecorator(story => (
    <div className="container mt-3">{story()}</div>
))

// Disable Chromatic for the non-GitHub popovers since they are mostly the same

const services = ['github', 'gitlab', 'phabricator', 'bitbucketServer'] as const

for (const serviceType of services) {
    add(
        `${serviceType}`,
        () => (
            <WebStory>
                {() => {
                    const targetID = `view-on-${serviceType}`
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
                                id={targetID}
                                onClick={() => setOpen(isOpen => !isOpen)}
                            >
                                View on {serviceType}
                            </button>
                            <InstallBrowserExtensionPopover
                                url=""
                                serviceType={serviceType}
                                onClose={onClose}
                                onRejection={onRejection}
                                onClickInstall={onClickInstall}
                                targetID={targetID}
                                isOpen={open}
                                toggle={noop}
                            />
                        </>
                    )
                }}
            </WebStory>
        ),
        {
            chromatic: {
                disable: serviceType !== 'github',
            },
        }
    )
}
