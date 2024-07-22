import React, { useCallback } from 'react'

import type { Meta } from '@storybook/react'

import { BrandedStory } from '../../../stories/BrandedStory'

import { Input, InputDescription, InputElement, InputErrorMessage, InputStatus, Label } from './Input'

const Story: Meta = {
    title: 'wildcard/Input',

    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],

    parameters: {
        component: Input,

        design: {
            type: 'figma',
            name: 'Figma',
            url: 'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Wildcard-Design-System?node-id=875%3A797',
        },
    },
}

export default Story

export const Simple = () => {
    const [selected, setSelected] = React.useState('')

    const handleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setSelected(event.target.value)
    }, [])

    return (
        <>
            <Input label="Input raw" value={selected} onChange={handleChange} />
            <Input
                value={selected}
                label="Input valid"
                onChange={handleChange}
                message="random message"
                status="valid"
                disabled={false}
                placeholder="testing this one"
            />
            <Input
                value={selected}
                label="Input loading"
                onChange={handleChange}
                message="random message"
                status="loading"
                placeholder="loading status input"
            />
            <Input
                value={selected}
                label="Input error"
                onChange={handleChange}
                error="An error message that can contain `code` or other **Markdown** _formatting_. [Learn more](https://sourcegraph.com/docs)"
                status="error"
                placeholder="error status input"
            />
            <Input
                value={selected}
                label="Disabled input"
                onChange={handleChange}
                message="random message"
                disabled={true}
                placeholder="disable status input"
            />

            <Input
                value={selected}
                label="Input small"
                onChange={handleChange}
                message="random message"
                status="valid"
                disabled={false}
                placeholder="testing this one"
                variant="small"
            />

            <section>
                <Label htmlFor="customInput">Custom label layout</Label>
                <InputElement
                    id="customInput"
                    placeholder="Field with custom label layout"
                    status={InputStatus.error}
                />
                <InputErrorMessage message="Input custom error message" className="mt-2" />
                <InputDescription className="mt-2">
                    <ul>
                        <li>Hint: you can use regular expressions within each of the available filters</li>
                        <li>
                            Datapoints will be automatically backfilled using the list of repositories resulting from
                            today’s search. Future data points will use the list refreshed for every snapshot.
                        </li>
                    </ul>
                </InputDescription>
            </section>
        </>
    )
}
