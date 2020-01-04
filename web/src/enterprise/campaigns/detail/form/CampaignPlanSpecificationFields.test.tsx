import React from 'react'
import { CampaignPlanSpecificationFields, MANUAL_CAMPAIGN_TYPE } from './CampaignPlanSpecificationFields'
import { createRenderer } from 'react-test-renderer/shallow'

jest.mock('../../../../settings/MonacoSettingsEditor', () => ({ MonacoSettingsEditor: 'MonacoSettingsEditor' }))

const PROPS = {
    onChange: () => undefined,
    isLightTheme: true,
}

describe('CampaignPlanSpecificationFields', () => {
    describe('manual type', () => {
        test('editable', () =>
            expect(
                createRenderer().render(
                    <CampaignPlanSpecificationFields
                        {...PROPS}
                        value={{ type: MANUAL_CAMPAIGN_TYPE, arguments: '' }}
                        readOnly={false}
                    />
                )
            ).toMatchSnapshot())

        test('read-only', () =>
            expect(
                createRenderer().render(
                    <CampaignPlanSpecificationFields
                        {...PROPS}
                        value={{ type: MANUAL_CAMPAIGN_TYPE, arguments: '' }}
                        readOnly={true}
                    />
                )
            ).toMatchSnapshot())
    })

    describe('non-manual type', () => {
        test('editable', () =>
            expect(
                createRenderer().render(
                    <CampaignPlanSpecificationFields
                        {...PROPS}
                        value={{ type: 'comby', arguments: '' }}
                        readOnly={false}
                    />
                )
            ).toMatchSnapshot())

        test('read-only', () =>
            expect(
                createRenderer().render(
                    <CampaignPlanSpecificationFields
                        {...PROPS}
                        value={{ type: 'comby', arguments: '' }}
                        readOnly={true}
                    />
                )
            ).toMatchSnapshot())
    })
})
