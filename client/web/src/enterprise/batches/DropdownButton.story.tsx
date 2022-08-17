import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { Action, DropdownButton } from './DropdownButton'

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

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/batches/DropdownButton',
    decorators: [decorator],
    argTypes: {
        disabled: {
            control: { type: 'boolean' },
            defaultValue: false,
        },
    },
}

export default config

export const NoActions: Story = args => <WebStory>{() => <DropdownButton actions={[]} {...args} />}</WebStory>
NoActions.argTypes = {
    disabled: {
        table: {
            disable: true,
        },
    },
}

NoActions.storyName = 'No actions'

export const SingleAction: Story = args => <WebStory>{() => <DropdownButton actions={[action]} {...args} />}</WebStory>

SingleAction.storyName = 'Single action'

export const MultipleActionsWithoutDefault: Story = args => (
    <WebStory>{() => <DropdownButton actions={[action, disabledAction, experimentalAction]} {...args} />}</WebStory>
)

MultipleActionsWithoutDefault.storyName = 'Multiple actions without default'

export const MultipleActionsWithDefault: Story = args => (
    <WebStory>
        {() => <DropdownButton actions={[action, disabledAction, experimentalAction]} defaultAction={0} {...args} />}
    </WebStory>
)

MultipleActionsWithDefault.storyName = 'Multiple actions with default'
