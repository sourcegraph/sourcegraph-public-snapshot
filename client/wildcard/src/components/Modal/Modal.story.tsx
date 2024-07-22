import type { Meta, StoryFn } from '@storybook/react'

import { H1 } from '..'
import { BrandedStory } from '../../stories/BrandedStory'

import { Modal } from '.'

const config: Meta = {
    title: 'wildcard/Modal',
    component: Modal,

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
}

export default config

export const Default: StoryFn = () => (
    <Modal aria-label="Welcome message">
        <H1>Hello world!</H1>
    </Modal>
)

Default.parameters = {
    component: Modal,

    design: [
        {
            type: 'figma',
            name: 'Figma Light',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=15122%3A40371',
        },
        {
            type: 'figma',
            name: 'Figma Dark',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=15122%3A40506',
        },
    ],
}

export const PositionCentered: StoryFn = () => (
    <Modal position="center" aria-label="Welcome message">
        <H1>Hello world!</H1>
    </Modal>
)

export const PositionFull: StoryFn = () => (
    <Modal position="full" aria-label="Welcome message">
        <H1>Hello world!</H1>
    </Modal>
)
