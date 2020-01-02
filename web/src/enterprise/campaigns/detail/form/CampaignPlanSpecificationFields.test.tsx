import React from 'react'
import sinon from 'sinon'
import { CampaignPlanSpecificationFields, MANUAL_CAMPAIGN_TYPE } from './CampaignPlanSpecificationFields'
import { createRenderer } from 'react-test-renderer/shallow'
import renderer, { act } from 'react-test-renderer'

jest.mock('../../../../settings/MonacoSettingsEditor', () => ({ MonacoSettingsEditor: 'MonacoSettingsEditor' }))

const PROPS = {
    onChange: () => undefined,
    isLightTheme: true,
}

describe('CampaignPlanSpecificationFields', () => {
    test('has initial value and calls onChange', () => {
        const onChange = sinon.spy()
        const component = renderer.create(
            <CampaignPlanSpecificationFields {...PROPS} value={undefined} onChange={onChange} />
        )
        act(() => undefined) // eslint-disable-line @typescript-eslint/no-floating-promises
        expect(component).toMatchSnapshot()

        expect(onChange.calledOnce).toBe(true)
        expect(onChange.firstCall.args[0].type).toBe('comby')
        expect(onChange.firstCall.args[0].arguments).toMatch(/scopeQuery/)
    })

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
