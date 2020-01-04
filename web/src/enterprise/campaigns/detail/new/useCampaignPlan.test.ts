import * as GQL from '../../../../../../shared/src/graphql/schema'
import { renderHook } from '@testing-library/react-hooks'
import { useCampaignPlan } from './useCampaignPlan'
import { of } from 'rxjs'

// Causes false positive on act().
/* eslint-disable @typescript-eslint/no-floating-promises */

const CAMPAIGN_PLAN_1 = {
    __typename: 'CampaignPlan',
    status: { state: GQL.BackgroundProcessState.COMPLETED },
} as GQL.ICampaignPlan

describe('useCampaignPlan', () => {
    test('undefined specOrID', () => {
        const { result } = renderHook(() => useCampaignPlan(undefined))
        expect(result.current).toEqual([undefined, false])
    })

    test('existing plan ID', () => {
        const { result } = renderHook(() => useCampaignPlan('p', () => of(CAMPAIGN_PLAN_1)))
        expect(result.current).toEqual([CAMPAIGN_PLAN_1, false])
    })

    test('plan spec', () => {
        const { result } = renderHook(() => useCampaignPlan({ type: 't', arguments: '{}' }, () => of(CAMPAIGN_PLAN_1)))
        expect(result.current).toEqual([CAMPAIGN_PLAN_1, false])
    })

    test('error', () => {
        const error = new Error('x')
        const { result } = renderHook(() => useCampaignPlan('p', () => of(error)))
        expect(result.current).toEqual([error, false])
    })
})
