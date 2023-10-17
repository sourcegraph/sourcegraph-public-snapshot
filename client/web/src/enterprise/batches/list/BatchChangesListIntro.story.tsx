import type { Meta, Decorator, StoryFn } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangesListIntro } from './BatchChangesListIntro'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

enum LicensingState {
    Licensed = 'Licensed',
    Unlicensed = 'Unlicensed',
    Loading = 'Loading',
}

const config: Meta = {
    title: 'web/batches/list/BatchChangesListIntro',
    decorators: [decorator],
    argTypes: {
        licensed: {
            control: { type: 'radio', options: LicensingState },
        },
        state: {
            table: {
                disable: true,
            },
        },
    },
}

export default config

function stateToInput(state: LicensingState): boolean | undefined {
    switch (state) {
        case LicensingState.Licensed:
            return true
        case LicensingState.Unlicensed:
            return false
        default:
            return undefined
    }
}

const Template: StoryFn = ({ state, ...args }) => (
    <WebStory>
        {() => <BatchChangesListIntro viewerIsAdmin={false} isLicensed={stateToInput(args.licensed)} />}
    </WebStory>
)

export const Licensed = Template.bind({})
Licensed.args = { state: LicensingState.Licensed }
Licensed.argTypes = {
    licensed: {},
}

export const Unlicensed = Template.bind({})
Unlicensed.args = { state: LicensingState.Unlicensed }
Unlicensed.argTypes = {
    licensed: {},
}

export const Loading = Template.bind({})
Loading.args = { state: LicensingState.Loading }
Loading.argTypes = {
    licensed: {},
}
