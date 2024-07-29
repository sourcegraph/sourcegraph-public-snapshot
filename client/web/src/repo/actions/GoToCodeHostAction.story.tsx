/* eslint-disable react-hooks/rules-of-hooks */
import { useEffect, useState } from 'react'

import { mdiGithub, mdiGitlab, mdiBitbucket } from '@mdi/js'
import type { Meta, StoryFn, Decorator } from '@storybook/react'

import { PhabricatorIcon } from '@sourcegraph/shared/src/components/icons'
import { Button, Popover, PopoverTrigger, Icon } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'
import { ExternalServiceKind } from '../../graphql-operations'

const decorator: Decorator = story => <div className="container mt-3">{story()}</div>

const config: Meta = {
    title: 'web/repo/actions/InstallBrowserExtensionPopover',
    decorators: [decorator],
}

export default config

export const GitHub: StoryFn = () => (
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
                <Popover isOpen={open} onOpenChange={event => setOpen(event.isOpen)}>
                    <PopoverTrigger as={Button} id={targetID} aria-label="Github">
                        <Icon aria-hidden="true" svgPath={mdiGithub} />
                    </PopoverTrigger>
                </Popover>
            )
        }}
    </WebStory>
)

GitHub.storyName = 'GitHub'

export const GitLab: StoryFn = () => (
    <WebStory>
        {() => {
            const serviceKind = ExternalServiceKind.GITLAB
            const targetID = `view-on-${serviceKind}`
            const [open, setOpen] = useState(false)
            useEffect(() => {
                setTimeout(() => setOpen(true), 0)
            }, [])
            return (
                <Popover isOpen={open} onOpenChange={event => setOpen(event.isOpen)}>
                    <PopoverTrigger as={Button} id={targetID} aria-label="Gitlab">
                        <Icon aria-hidden={true} svgPath={mdiGitlab} />
                    </PopoverTrigger>
                </Popover>
            )
        }}
    </WebStory>
)

GitLab.storyName = 'GitLab'

export const Phabricator: StoryFn = () => (
    <WebStory>
        {() => {
            const serviceKind = ExternalServiceKind.PHABRICATOR
            const targetID = `view-on-${serviceKind}`
            const [open, setOpen] = useState(false)
            useEffect(() => {
                setTimeout(() => setOpen(true), 0)
            }, [])
            return (
                <Popover isOpen={open} onOpenChange={event => setOpen(event.isOpen)}>
                    <PopoverTrigger as={Button} id={targetID} aria-label="Phabricator">
                        <Icon as={PhabricatorIcon} aria-hidden={true} />
                    </PopoverTrigger>
                </Popover>
            )
        }}
    </WebStory>
)

export const BitbucketServer: StoryFn = () => (
    <WebStory>
        {() => {
            const serviceKind = ExternalServiceKind.BITBUCKETSERVER
            const targetID = `view-on-${serviceKind}`
            const [open, setOpen] = useState(false)
            useEffect(() => {
                setTimeout(() => setOpen(true), 0)
            }, [])
            return (
                <Popover isOpen={open} onOpenChange={event => setOpen(event.isOpen)}>
                    <PopoverTrigger as={Button} id={targetID} aria-label="Bitbucket">
                        <Icon aria-hidden={true} svgPath={mdiBitbucket} />
                    </PopoverTrigger>
                </Popover>
            )
        }}
    </WebStory>
)

BitbucketServer.storyName = 'Bitbucket server'
