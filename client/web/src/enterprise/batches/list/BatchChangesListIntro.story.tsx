import { radios } from '@storybook/addon-knobs'
import { Meta, DecoratorFn, Story } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangesListIntro } from './BatchChangesListIntro'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>
const config: Meta = {
    title: 'web/batches/list/BatchChangesListIntro',
    decorators: [decorator],
}

export default config

enum LicensingState {
    Licensed = 'Licensed',
    Unlicensed = 'Unlicensed',
    Loading = 'Loading',
}

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

const Template: Story<{ state: LicensingState }> = ({ state }) => (
    <WebStory>
        {() => <BatchChangesListIntro isLicensed={stateToInput(radios('licensed', LicensingState, state))} />}
    </WebStory>
)

export const Licensed = Template.bind({})
Licensed.args = { state: LicensingState.Licensed }

export const Unlicensed = Template.bind({})
Unlicensed.args = { state: LicensingState.Unlicensed }

export const Loading = Template.bind({})
Loading.args = { state: LicensingState.Loading }
