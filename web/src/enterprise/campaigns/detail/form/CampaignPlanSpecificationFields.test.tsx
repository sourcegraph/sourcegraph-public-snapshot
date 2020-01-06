import React from 'react'
import { CampaignPlanSpecificationFields } from './CampaignPlanSpecificationFields'
import { createRenderer } from 'react-test-renderer/shallow'

jest.mock('../../../../settings/MonacoSettingsEditor', () => ({ MonacoSettingsEditor: 'MonacoSettingsEditor' }))

const PROPS = {
    onTypeChange: () => undefined,
    argumentsJSONC: undefined,
    onArgumentsJSONCChange: () => undefined,
    isLightTheme: true,
}

describe('CampaignPlanSpecificationFields', () => {
    describe('manual type', () => {
        test('editable', () =>
            expect(
                createRenderer().render(
                    <CampaignPlanSpecificationFields {...PROPS} type={undefined} readOnly={false} />
                )
            ).toMatchSnapshot())

        test('read-only', () =>
            expect(
                createRenderer().render(<CampaignPlanSpecificationFields {...PROPS} type={undefined} readOnly={true} />)
            ).toMatchSnapshot())
    })

    describe('non-manual type', () => {
        test('editable', () =>
            expect(
                createRenderer().render(<CampaignPlanSpecificationFields {...PROPS} type="comby" readOnly={false} />)
            ).toMatchSnapshot())

        test('read-only', () =>
            expect(
                createRenderer().render(<CampaignPlanSpecificationFields {...PROPS} type="comby" readOnly={true} />)
            ).toMatchSnapshot())
    })
})
