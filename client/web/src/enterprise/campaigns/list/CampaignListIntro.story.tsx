import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { CampaignListIntro } from './CampaignListIntro'

const { add } = storiesOf('web/campaigns/CampaignListIntro', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Intro', () => (
    <EnterpriseWebStory>
        {() => (
            <CampaignListIntro licensed={stateToInput(radios('licensed', LicensingState, LicensingState.Loading))} />
        )}
    </EnterpriseWebStory>
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
