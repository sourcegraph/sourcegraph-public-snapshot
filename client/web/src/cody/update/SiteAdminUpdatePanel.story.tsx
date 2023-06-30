import '@sourcegraph/branded'

import { Meta, StoryObj } from '@storybook/react'

import { BrandedStory } from '@sourcegraph/wildcard/src/stories'

import 'storybook-addon-designs'

import { Container } from '@sourcegraph/wildcard'

import { SiteAdminUpdatePanelFrame, SiteAdminUpdatePanelProps } from './SiteAdminUpdatePanel'

const meta: Meta<SiteAdminUpdatePanelProps> = {
    title: 'cody-ui/Updater/Admin Page',
    decorators: [
        Story => (
            <div>
                We just render the update info component in a container. All states defined in the Content storybook
                will show here accordingly.
                <Container className="mt-3 pb-4" style={{ border: '1px dashed gray' }}>
                    <Container className="container mt-0 pb-4">
                        <Story />
                    </Container>
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
        update: {
            stage: 'IDLE',
            hasNewVersion: true,
            description: 'The quick brown fox jumped over the lazy dog.',
            startInstall: () => {},
        },
    },
    component: SiteAdminUpdatePanelFrame,
}

export default meta

type Story = StoryObj<SiteAdminUpdatePanelProps>

export const Frame: Story = {}
