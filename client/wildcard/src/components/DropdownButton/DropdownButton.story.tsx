import { Meta, Story } from '@storybook/react'

import { BrandedStory } from '../../stories'

import { DropdownButtonAction, DropdownButton } from './DropdownButton'

// eslint-disable-next-line @typescript-eslint/require-await
const onTrigger = async (onDone: () => void) => onDone()

const action: DropdownButtonAction = {
    type: 'action-type',
    buttonLabel: 'Action',
    dropdownTitle: 'Action',
    dropdownDescription: 'Perform an action',
    onTrigger,
}

const disabledAction: DropdownButtonAction = {
    type: 'disabled-action-type',
    buttonLabel: 'Disabled action',
    disabled: true,
    dropdownTitle: 'Disabled action',
    dropdownDescription: 'Perform an action, if only this were enabled',
    onTrigger,
}

const experimentalAction: DropdownButtonAction = {
    type: 'experimental-action-type',
    buttonLabel: 'Experimental action',
    dropdownTitle: 'Experimental action',
    dropdownDescription: 'Perform a super cool action that might explode',
    onTrigger,
    experimental: true,
}

const config: Meta = {
    title: 'wildcard/DropdownButton',
    decorators: [story => <BrandedStory>{() => <div className="container mt-3">{story()}</div>}</BrandedStory>],
    parameters: {
        chromatic: {
            enableDarkMode: true,
            disableSnapshot: false,
        },
    },
    argTypes: {
        disabled: {
            control: { type: 'boolean' },
            defaultValue: false,
        },
    },
}

export default config

export const NoActions: Story = args => <DropdownButton actions={[]} {...args} />
NoActions.argTypes = {
    disabled: {
        table: {
            disable: true,
        },
    },
}

NoActions.storyName = 'No actions'

export const SingleAction: Story = args => <DropdownButton actions={[action]} {...args} />

SingleAction.storyName = 'Single action'

export const MultipleActionsWithoutDefault: Story = args => (
    <DropdownButton actions={[action, disabledAction, experimentalAction]} {...args} />
)

MultipleActionsWithoutDefault.storyName = 'Multiple actions without default'

export const MultipleActionsWithDefault: Story = args => (
    <DropdownButton actions={[action, disabledAction, experimentalAction]} defaultAction={0} {...args} />
)

MultipleActionsWithDefault.storyName = 'Multiple actions with default'
