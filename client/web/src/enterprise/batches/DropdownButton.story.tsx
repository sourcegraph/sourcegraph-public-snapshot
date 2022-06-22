import { boolean, select } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { Action, DropdownButton, Props } from './DropdownButton'

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
    disabled: true,
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

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/DropdownButton',
    decorators: [decorator],
}

export default config

export const NoActions: Story = () => <WebStory>{() => <DropdownButton actions={[]} {...commonKnobs()} />}</WebStory>

NoActions.storyName = 'No actions'

export const SingleAction: Story = () => (
    <WebStory>{() => <DropdownButton actions={[action]} {...commonKnobs()} />}</WebStory>
)

SingleAction.storyName = 'Single action'

export const MultipleActionsWithoutDefault: Story = () => (
    <WebStory>
        {() => <DropdownButton actions={[action, disabledAction, experimentalAction]} {...commonKnobs()} />}
    </WebStory>
)

MultipleActionsWithoutDefault.storyName = 'Multiple actions without default'

export const MultipleActionsWithDefault: Story = () => (
    <WebStory>
        {() => (
            <DropdownButton
                actions={[action, disabledAction, experimentalAction]}
                defaultAction={0}
                {...commonKnobs()}
            />
        )}
    </WebStory>
)

MultipleActionsWithDefault.storyName = 'Multiple actions with default'
