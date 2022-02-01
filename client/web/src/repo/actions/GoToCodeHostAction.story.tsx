import { action } from '@storybook/addon-actions'
import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import BitbucketIcon from 'mdi-react/BitbucketIcon'
import GithubIcon from 'mdi-react/GithubIcon'
import GitlabIcon from 'mdi-react/GitlabIcon'
import React, { useEffect, useState } from 'react'

import { PhabricatorIcon } from '@sourcegraph/shared/src/components/icons'
import { ExternalServiceKind } from '@sourcegraph/shared/src/schema'
import { Button } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'

import { InstallBrowserExtensionPopover } from './InstallBrowserExtensionPopover'

const onClose = action('onClose')
const onReject = action('onReject')
const onInstall = action('onInstall')

const { add } = storiesOf('web/repo/actions/InstallBrowserExtensionPopover', module).addDecorator(story => (
    <div className="container mt-3">{story()}</div>
))

add('GitHub', () => (
    <WebStory>
        {() => {
            const serviceKind = ExternalServiceKind.GITHUB
            const targetID = `view-on-${serviceKind}`
            const [open, setOpen] = useState(false)
            // The popover cannot be open on initial render
            // since the target element isn't in the DOM yet
            useEffect(() => {
                setTimeout(() => setOpen(true), 0)
            }, [])
            return (
                <>
                    <Button id={targetID} onClick={() => setOpen(isOpen => !isOpen)}>
                        <GithubIcon className="icon-inline" />
                    </Button>
                    <InstallBrowserExtensionPopover
                        url=""
                        serviceKind={serviceKind}
                        onClose={onClose}
                        onReject={onReject}
                        onInstall={onInstall}
                        targetID={targetID}
                        isOpen={open}
                        onToggle={noop}
                    />
                </>
            )
        }}
    </WebStory>
))

// Disable Chromatic for the non-GitHub popovers since they are mostly the same

add(
    'GitLab',
    () => (
        <WebStory>
            {() => {
                const serviceKind = ExternalServiceKind.GITLAB
                const targetID = `view-on-${serviceKind}`
                const [open, setOpen] = useState(false)
                useEffect(() => {
                    setTimeout(() => setOpen(true), 0)
                }, [])
                return (
                    <>
                        <Button id={targetID} onClick={() => setOpen(isOpen => !isOpen)}>
                            <GitlabIcon className="icon-inline" />
                        </Button>
                        <InstallBrowserExtensionPopover
                            url=""
                            serviceKind={serviceKind}
                            onClose={onClose}
                            onReject={onReject}
                            onInstall={onInstall}
                            targetID={targetID}
                            isOpen={open}
                            onToggle={noop}
                        />
                    </>
                )
            }}
        </WebStory>
    ),
    {
        chromatic: {
            disable: true,
        },
    }
)

add(
    'Phabricator',
    () => (
        <WebStory>
            {() => {
                const serviceKind = ExternalServiceKind.PHABRICATOR
                const targetID = `view-on-${serviceKind}`
                const [open, setOpen] = useState(false)
                useEffect(() => {
                    setTimeout(() => setOpen(true), 0)
                }, [])
                return (
                    <>
                        <Button id={targetID} onClick={() => setOpen(isOpen => !isOpen)}>
                            <PhabricatorIcon className="icon-inline" />
                        </Button>
                        <InstallBrowserExtensionPopover
                            url=""
                            serviceKind={serviceKind}
                            onClose={onClose}
                            onReject={onReject}
                            onInstall={onInstall}
                            targetID={targetID}
                            isOpen={open}
                            onToggle={noop}
                        />
                    </>
                )
            }}
        </WebStory>
    ),
    {
        chromatic: {
            disable: true,
        },
    }
)

add(
    'Bitbucket server',
    () => (
        <WebStory>
            {() => {
                const serviceKind = ExternalServiceKind.BITBUCKETSERVER
                const targetID = `view-on-${serviceKind}`
                const [open, setOpen] = useState(false)
                useEffect(() => {
                    setTimeout(() => setOpen(true), 0)
                }, [])
                return (
                    <>
                        <Button id={targetID} onClick={() => setOpen(isOpen => !isOpen)}>
                            <BitbucketIcon className="icon-inline" />
                        </Button>

                        <InstallBrowserExtensionPopover
                            url=""
                            serviceKind={serviceKind}
                            onClose={onClose}
                            onReject={onReject}
                            onInstall={onInstall}
                            targetID={targetID}
                            isOpen={open}
                            onToggle={noop}
                        />
                    </>
                )
            }}
        </WebStory>
    ),
    {
        chromatic: {
            disable: true,
        },
    }
)
