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

export const Licensed: Story = () => (
    <WebStory>
        {() => (
            <BatchChangesListIntro
                isLicensed={stateToInput(radios('licensed', LicensingState, LicensingState.Licensed))}
            />
        )}
    </WebStory>
)

export const Unlicensed: Story = () => (
    <WebStory>
        {() => (
            <BatchChangesListIntro
                isLicensed={stateToInput(radios('licensed', LicensingState, LicensingState.Unlicensed))}
            />
        )}
    </WebStory>
)

export const Loading: Story = () => (
    <WebStory>
        {() => (
            <BatchChangesListIntro
                isLicensed={stateToInput(radios('licensed', LicensingState, LicensingState.Loading))}
            />
        )}
    </WebStory>
)
