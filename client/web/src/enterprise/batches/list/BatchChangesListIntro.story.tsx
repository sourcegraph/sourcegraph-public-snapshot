import { radios } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { BatchChangesListIntro } from './BatchChangesListIntro'

const { add } = storiesOf('web/batches/list/BatchChangesListIntro', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

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

for (const state of Object.values(LicensingState)) {
    add(state, () => (
        <WebStory>
            {() => <BatchChangesListIntro isLicensed={stateToInput(radios('licensed', LicensingState, state))} />}
        </WebStory>
    ))
}
