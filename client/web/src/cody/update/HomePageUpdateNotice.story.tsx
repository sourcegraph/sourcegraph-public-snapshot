import '@sourcegraph/branded'

import { Meta, StoryObj } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import 'storybook-addon-designs'

import { Container } from '@sourcegraph/wildcard'

import { HomePageUpdateNoticeFrame, HomePageUpdateNoticeFrameProps } from './HomePageUpdateNotice'

const meta: Meta<HomePageUpdateNoticeFrameProps> = {
    title: 'cody-ui/Updater/HomePage',
    decorators: [
        Story => (
            <div>
                We just render the update info component in a container. All states defined in the Content storybook
                will show here accordingly.
                <div className="mt-3">
                    <Story />
                </div>
            </div>
        ),
        Story => (
            <BrandedStory>
                {() => (
                    <div className="container mt-3 pb-3">
                        <Story />
                    </div>
                )}
            </BrandedStory>
        ),
    ],
    component: HomePageUpdateNoticeFrame,
}

export default meta

type Story = StoryObj<HomePageUpdateNoticeFrameProps>

export const NoUpdates: Story = {
    args: {
        update: {
            stage: 'IDLE',
            hasNewVersion: false,
        },
    },
    decorators: [
        Story => (
            <div>
                Nothing is supposed to be rendered inside the box below:
                <Container className="mt-3 p-1" style={{ border: '1px dashed gray' }}>
                    <Story />
                </Container>
            </div>
        ),
    ],
}

export const HasUpdate: Story = {
    args: {
        update: {
            stage: 'IDLE',
            hasNewVersion: true,
            version: '1.0.0',
            newVersion: '2.0.0',
            description: 'The quick brown fox jumped over the lazy dog.',
            startInstall: () => {},
        },
    },
    decorators: [
        Story => (
            <div>
                <Container className="mt-3 pb-2" style={{ border: '1px dashed gray' }}>
                    <Story />
                </Container>
            </div>
        ),
    ],
}
