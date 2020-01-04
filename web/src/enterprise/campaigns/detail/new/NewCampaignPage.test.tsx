import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import renderer, { act } from 'react-test-renderer'
import { NewCampaignPage } from './NewCampaignPage'
import * as H from 'history'
import { CampaignTabs } from '../CampaignTabs'
import { NewCampaignForm, NewCampaignFormData } from './NewCampaignForm'

// Ignore false positive on act().
/* eslint-disable @typescript-eslint/no-floating-promises */

jest.mock('../../../../settings/MonacoSettingsEditor', () => ({ MonacoSettingsEditor: 'MonacoSettingsEditor' }))
jest.mock('../CampaignTabs', () => ({ CampaignTabs: 'CampaignTabs' }))

const history = H.createMemoryHistory()

const PROPS = {
    authenticatedUser: { id: 'u', username: 'alice', avatarURL: 'https://example.com/alice' },
    history,
    location: history.location,
    isLightTheme: true,
}

const CAMPAIGN_PLAN_1 = {
    __typename: 'CampaignPlan',
    status: { state: GQL.BackgroundProcessState.COMPLETED },
} as GQL.ICampaignPlan

describe('NewCampaignPage', () => {
    test('existing plan', () => {
        const location = H.createLocation({ search: 'plan=p' })
        const component = renderer.create(
            <NewCampaignPage {...PROPS} location={location} _useCampaignPlan={() => [undefined, true]} />
        )
        expect(
            component.root.findAll(
                node => node.type === 'button' && node.props.className.includes('e2e-preview-campaign')
            ).length
        ).toBe(0)
        const createButton = component.root.find(
            node => node.type === 'button' && node.props.className.includes('e2e-create-campaign')
        )
        expect(createButton.props.disabled).toBeFalsy()
    })

    test('clear plan when type changes', () => {
        const component = renderer.create(
            <NewCampaignPage
                {...PROPS}
                _useCampaignPlan={specOrID => (specOrID ? [CAMPAIGN_PLAN_1, false] : [undefined, false])}
            />
        )
        const previewButton = component.root.find(
            node => node.type === 'button' && node.props.className.includes('e2e-preview-campaign')
        )
        act(() => previewButton.props.onClick())
        expect(component.root.findByType(CampaignTabs)).toBeDefined()

        const form = component.root.findByType(NewCampaignForm)
        const value = form.props.value
        act(() => form.props.onChange({ ...value, plan: { type: 't', arguments: '{}' } } as NewCampaignFormData))
        expect(component.root.findAllByType(CampaignTabs).length).toBe(0)
    })
})
