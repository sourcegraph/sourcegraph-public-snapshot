import React from 'react'
import renderer from 'react-test-renderer'
import { NewCampaignForm } from './NewCampaignForm'
import {
    CampaignPlanSpecificationFields,
    DEFAULT_CAMPAIGN_PLAN_SPECIFICATION_FORM_DATA,
} from '../form/CampaignPlanSpecificationFields'

jest.mock('../form/CampaignPlanSpecificationFields', () => ({
    CampaignPlanSpecificationFields: 'CampaignPlanSpecificationFields',
}))
jest.mock('../form/CampaignTitleField', () => ({ CampaignTitleField: 'CampaignTitleField' }))
jest.mock('../form/CampaignDescriptionField', () => ({ CampaignDescriptionField: 'CampaignDescriptionField' }))

const PROPS = {
    onChange: () => undefined,
    isLightTheme: true,
}

describe('NewCampaignForm', () => {
    test('existing plan', () => {
        const component = renderer.create(
            <NewCampaignForm {...PROPS} value={{ name: '', description: '', plan: 'p' }} />
        )
        expect(component.root.findAllByType(CampaignPlanSpecificationFields).length).toBe(0)
    })

    test('no plan', () => {
        const component = renderer.create(
            <NewCampaignForm
                {...PROPS}
                value={{ name: '', description: '', plan: DEFAULT_CAMPAIGN_PLAN_SPECIFICATION_FORM_DATA }}
            />
        )
        expect(component.root.findAllByType(CampaignPlanSpecificationFields).length).toBe(1)
    })
})
