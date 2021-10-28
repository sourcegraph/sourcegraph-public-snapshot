import { DecoratorFn, Meta, Story } from '@storybook/react'
import React, { useCallback } from 'react'

import { BrandedStory } from '../BrandedStory'

import { Tooltip } from './Tooltip'

const PinnedTooltip: React.FunctionComponent = () => {
    const clickElement = useCallback((element: HTMLElement | null) => {
        if (element) {
            element.click()
        }
    }, [])
    return (
        <>
            <Tooltip />
            <span data-tooltip="My tooltip" ref={clickElement}>
                Example
            </span>
            <p>
                <small>
                    (A pinned tooltip is shown when the target element is rendered, without any user interaction
                    needed.)
                </small>
            </p>
        </>
    )
}

const decorator: DecoratorFn = story => <BrandedStory>{() => <div className="p-5">{story()}</div>}</BrandedStory>
const config: Meta = {
    title: 'branded/Tooltip',
    decorators: [decorator],
}

export default config

export const Hover: Story = () => (
    <>
        <Tooltip />
        <p>
            You can <strong data-tooltip="Tooltip 1">hover me</strong> or <strong data-tooltip="Tooltip 2">me</strong>.
        </p>
    </>
)

Hover.parameters = {
    chromatic: {
        disable: true,
    },
}

export const Pinned: Story = () => <PinnedTooltip />

Pinned.parameters = {
    chromatic: {
        // Chromatic pauses CSS animations by default and resets them to their initial state
        pauseAnimationAtEnd: true,
    },
}
