import { Meta } from '@storybook/react'
import React, { useCallback } from 'react'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { Grid } from '../../Grid'
import { BaseControlInputProps } from '../internal/BaseControlInput'

import { Checkbox } from './Checkbox'

const config: Meta = {
    title: 'wildcard/Checkbox',

    decorators: [
        story => (
            <BrandedStory styles={webStyles}>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>
        ),
    ],

    parameters: {
        component: Checkbox,
        design: {
            type: 'figma',
            name: 'Figma',
            url:
                'https://www.figma.com/file/NIsN34NH7lPu04olBzddTw/Design-Refresh-Systemization-source-of-truth?node-id=908%3A1353',
        },
    },
}

// eslint-disable-next-line import/no-default-export
export default config

const BaseCheckbox = (props: Pick<BaseControlInputProps, 'id' | 'isValid' | 'disabled'>) => {
    const [isChecked, setChecked] = React.useState(false)

    const handleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(event => {
        setChecked(event.target.checked)
    }, [])

    return (
        <Checkbox
            name="example-1"
            value="first"
            checked={isChecked}
            onChange={handleChange}
            label="First"
            message="Hello world!"
            {...props}
        />
    )
}

export const CheckboxExamples: React.FunctionComponent = () => (
    <>
        <h1>Checkbox</h1>
        <Grid columnCount={4}>
            <div>
                <h2>Standard</h2>
                <BaseCheckbox id="standard-example" />
            </div>
            <div>
                <h2>Valid</h2>
                <BaseCheckbox id="valid-example" isValid={true} />
            </div>
            <div>
                <h2>Invalid</h2>
                <BaseCheckbox id="invalid-example" isValid={false} />
            </div>
            <div>
                <h2>Disabled</h2>
                <BaseCheckbox id="disabled-example" disabled={true} />
            </div>
        </Grid>
    </>
)
