import '@sourcegraph/branded'

import type { Meta, StoryObj } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import '@storybook/addon-designs'

import { Container } from '@sourcegraph/wildcard'

import { type UpdateGlobalNav, UpdateGlobalNavFrame, type UpdateGlobalNavFrameProps } from './UpdateGlobalNav'

const meta: Meta<UpdateGlobalNavFrameProps> = {
    title: 'cody-ui/Updater/Global Nav',
    decorators: [
        Story => (
            <div>
                We just render the update info component in a container. All states defined in the Content storybook
                will show here accordingly.
                <Container className="mt-3 pb-4" style={{ border: '1px dashed gray' }}>
                    <Story />
                </Container>
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
    args: {
        details: {
            stage: 'IDLE',
            hasNewVersion: true,
            newVersion: '1.2.3',
            description: 'The quick brown fox jumped over the lazy dog.',
            startInstall: () => {},
        },
    },
    component: UpdateGlobalNavFrame,
}

export default meta

type Story = StoryObj<typeof UpdateGlobalNav>

export const NoNewVersion: Story = {
    args: {
        details: {
            hasNewVersion: false,
        },
    },
}

export const HasNewVersion: Story = {}
