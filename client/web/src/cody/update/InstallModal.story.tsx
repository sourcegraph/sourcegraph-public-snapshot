import '@sourcegraph/branded'

import type { Meta, StoryObj } from '@storybook/react'

import '@storybook/addon-designs'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import { InstallModal, type InstallModalProps } from './InstallModal'

const meta: Meta<InstallModalProps> = {
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
    component: InstallModal,
}

export default meta

type Story = StoryObj<InstallModalProps>

export const Installing: Story = {}

export const InstallFailed: Story = {
    args: {
        details: {
            stage: 'ERROR',
            error: 'Something bad happened. Sorry.',
            hasNewVersion: true,
        },
    },
}
