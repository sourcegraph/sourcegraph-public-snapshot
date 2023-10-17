import '@sourcegraph/branded'

import type { Meta, StoryObj } from '@storybook/react'

import '@storybook/addon-designs'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { ChangelogModal, type ChangelogModalProps } from './ReviewAndInstallModal'

const meta: Meta<ChangelogModalProps> = {
    title: 'cody-ui/Updater/Dialogs',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3 pb-3">{story()}</div>}</BrandedStory>],
    args: {
        details: {
            stage: 'IDLE',
            hasNewVersion: true,
            newVersion: '1.2.3',
            description: 'The quick brown fox jumped over the lazy dog.',
            startInstall: () => {},
        },
    },
    component: ChangelogModal,
}

export default meta

type Story = StoryObj<ChangelogModalProps>

export const Changelog: Story = {}

export const ChangelogLongLog: Story = {
    args: {
        details: {
            stage: 'IDLE',
            hasNewVersion: true,
            newVersion: '9.8.7',
            description: `Nostrud adipisicing laborum aute minim tempor. Commodo
anim enim mollit in deserunt sit nostrud excepteur anim minim quis commodo irure
enim. Laborum cillum sint reprehenderit minim ad nostrud sint quis irure
adipisicing tempor commodo.

* This is markdown
* So we can do this

...

1. And this
2. and this

**Adipisicing consequat nisi dolore proident ipsum consequat cupidatat non magna
sunt sit eiusmod reprehenderit.** Mollit laboris eu nostrud commodo aliquip amet
amet sit. Aliquip adipisicing ut duis qui mollit est deserunt laboris. Officia
eiusmod pariatur qui consequat mollit cupidatat exercitation cupidatat
reprehenderit exercitation qui dolor.

Aute tempor ipsum deserunt cupidatat. Labore ad sunt ut ex. Ad fugiat duis
consequat et eu culpa est enim. Id magna minim duis nisi nisi. Do eu commodo do
aliquip laboris cupidatat. Qui proident pariatur in ea minim sint enim pariatur
et Lorem non non aliqua sunt.

Quis in non dolor mollit id ex id dolor ad. Do deserunt proident ullamco nostrud
do deserunt officia nisi et exercitation irure sunt eu do. Lorem sit eiusmod ut
consectetur in officia eu mollit amet laborum pariatur pariatur.`,
        },
    },
}
