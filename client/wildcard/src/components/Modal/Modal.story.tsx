import { Meta, Story } from '@storybook/react'
import React from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Modal } from '.'

const config: Meta = {
    title: 'wildcard/Modal',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Modal,
    },
}

export default config

export const Default: Story = () => (
    <Modal aria-label="Welcome message">
        <h1>Hello world!</h1>
    </Modal>
)

export const PositionCentered: Story = () => (
    <Modal position="center" aria-label="Welcome message">
        <h1>Hello world!</h1>
    </Modal>
)

export const PositionFull: Story = () => (
    <Modal position="full" aria-label="Welcome message">
        <h1>Hello world!</h1>
    </Modal>
)
