import { DecoratorFn, Meta, Story } from '@storybook/react'
import React, { useRef, useState } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Button } from '../Button'

import { Popover } from './Popover'

const decorator: DecoratorFn = story => (
    <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'wildcard/Popover',
    decorators: [decorator],
}

export default config

export const PopoverExample: Story = () => {
    const buttonReference = useRef<HTMLButtonElement>(null)

    const [isVisible, setVisibility] = useState(true)

    return (
        <div className="d-flex align-items-center">
            <Button ref={buttonReference} variant="secondary" outline={true}>
                Open
            </Button>

            <Popover isOpen={isVisible} target={buttonReference} onVisibilityChange={setVisibility}>
                <div className="p-3">
                    <h1>Hello world!</h1>
                </div>
            </Popover>
        </div>
    )
}
