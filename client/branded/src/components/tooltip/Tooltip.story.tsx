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

export default {
    title: 'branded/Tooltip',

    decorators: [story => <BrandedStory>{() => <div className="p-5">{story()}</div>}</BrandedStory>],
}

export const Hover = () => (
    <>
        <Tooltip />
        <p>
            You can <strong data-tooltip="Tooltip 1">hover me</strong> or <strong data-tooltip="Tooltip 2">me</strong>.
        </p>
    </>
)

Hover.story = {
    parameters: {
        chromatic: {
            disable: true,
        },
    },
}

export const Pinned = () => <PinnedTooltip />

Pinned.story = {
    parameters: {
        chromatic: {
            // Chromatic pauses CSS animations by default and resets them to their initial state
            pauseAnimationAtEnd: true,
        },
    },
}
