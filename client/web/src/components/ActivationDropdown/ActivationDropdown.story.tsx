import { DecoratorFn, Meta, Story } from '@storybook/react'
import * as H from 'history'

import { subtypeOf } from '@sourcegraph/common'

import { WebStory } from '../WebStory'

import { ActivationDropdown, ActivationDropdownProps } from './ActivationDropdown'
import { baseActivation } from './ActivationDropdown.fixtures'

const history = H.createMemoryHistory({ keyLength: 0 })
const commonProps = subtypeOf<Partial<ActivationDropdownProps>>()({
    alwaysShow: true,
    history,
    portal: false,
})

const decorator: DecoratorFn = story => (
    <WebStory>{() => <div className="container h-100 web-content">{story()}</div>}</WebStory>
)

const config: Meta = {
    title: 'web/ActivationDropdown',
    decorators: [decorator],
}

export default config

export const Loading: Story = () => <ActivationDropdown {...commonProps} activation={baseActivation()} />

export const _04Completed: Story = args => (
    <ActivationDropdown
        {...commonProps}
        activation={{
            ...baseActivation(),
            completed: {
                ...args,
            },
        }}
    />
)
_04Completed.argTypes = {
    ConnectedCodeHost: {
        control: { type: 'boolean' },
        defaultValue: false,
    },
    DidSearch: {
        control: { type: 'boolean' },
        defaultValue: false,
    },
    FoundReferences: {
        control: { type: 'boolean' },
        defaultValue: false,
    },
    EnabledSharing: {
        control: { type: 'boolean' },
        defaultValue: false,
    },
}

_04Completed.storyName = 'Progress 0/4 completed'
_04Completed.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
        chromatic: { viewports: [480] },
    },
}

export const _14Completed: Story = args => (
    <ActivationDropdown
        {...commonProps}
        activation={{
            ...baseActivation(),
            completed: {
                ...args,
            },
        }}
    />
)
_14Completed.argTypes = {
    ConnectedCodeHost: {
        control: { type: 'boolean' },
        defaultValue: true,
    },
    DidSearch: {
        control: { type: 'boolean' },
        defaultValue: false,
    },
    FoundReferences: {
        control: { type: 'boolean' },
        defaultValue: false,
    },
    EnabledSharing: {
        control: { type: 'boolean' },
        defaultValue: false,
    },
}

_14Completed.storyName = 'Progress 1/4 completed'
_14Completed.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
        chromatic: { viewports: [480] },
    },
}
