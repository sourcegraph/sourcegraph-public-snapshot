import { boolean, select } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../components/WebStory'

import { Action, DropdownButton, Props } from './DropdownButton'

const { add } = storiesOf('web/batches/DropdownButton', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

// eslint-disable-next-line @typescript-eslint/require-await
const onTrigger = async (onDone: () => void) => onDone()

const action: Action = {
    type: 'action-type',
    buttonLabel: 'Action',
    dropdownTitle: 'Action',
    dropdownDescription: 'Perform an action',
    onTrigger,
}

const disabledAction: Action = {
    type: 'disabled-action-type',
    buttonLabel: 'Disabled action',
    dropdownTitle: 'Disabled action',
    dropdownDescription: 'Perform an action, if only this were enabled',
    onTrigger,
}

const experimentalAction: Action = {
    type: 'experimental-action-type',
    buttonLabel: 'Experimental action',
    dropdownTitle: 'Experimental action',
    dropdownDescription: 'Perform a super cool action that might explode',
    onTrigger,
    experimental: true,
}

const commonKnobs: () => Pick<Props, 'disabled' | 'dropdownMenuPosition'> = () => ({
    disabled: boolean('Disabled', false),
    dropdownMenuPosition: select(
        'Dropdown menu position',
        {
            None: undefined,
            Left: 'left',
            Right: 'right',
        },
        undefined
    ),
})

add('No actions', () => <WebStory>{() => <DropdownButton actions={[]} {...commonKnobs()} />}</WebStory>)

add('Single action', () => <WebStory>{() => <DropdownButton actions={[action]} {...commonKnobs()} />}</WebStory>)

add('Multiple actions without default', () => (
    <WebStory>
        {() => <DropdownButton actions={[action, disabledAction, experimentalAction]} {...commonKnobs()} />}
    </WebStory>
))

add('Multiple actions with default', () => (
    <WebStory>
        {() => (
            <DropdownButton
                actions={[action, disabledAction, experimentalAction]}
                defaultAction={1}
                {...commonKnobs()}
            />
        )}
    </WebStory>
))
