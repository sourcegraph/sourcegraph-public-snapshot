import React from 'react'
import { CampaignDiffStat } from './CampaignDiffStat'
import { shallow } from 'enzyme'

describe('CampaignDiffStat', () => {
    test('for campaign', () =>
        expect(
            shallow(
                <CampaignDiffStat
                    campaign={{
                        diffStat: {
                            added: 888,
                            deleted: 777,
                            changed: 999,
                        },
                    }}
                    className="abc"
                />
            )
        ).toMatchSnapshot())
    test('hidden for empty campaign', () =>
        expect(
            shallow(
                <CampaignDiffStat
                    campaign={{
                        diffStat: {
                            added: 0,
                            deleted: 0,
                            changed: 0,
                        },
                    }}
                    className="abc"
                />
            )
        ).toMatchSnapshot())
    test('for patch set', () =>
        expect(
            shallow(
                <CampaignDiffStat
                    patchSet={{
                        diffStat: {
                            added: 888,
                            changed: 777,
                            deleted: 999,
                        },
                    }}
                    className="abc"
                />
            )
        ).toMatchSnapshot())
    test('hidden for empty patch set', () =>
        expect(
            shallow(
                <CampaignDiffStat
                    patchSet={{
                        diffStat: {
                            added: 0,
                            changed: 0,
                            deleted: 0,
                        },
                    }}
                    className="abc"
                />
            )
        ).toMatchSnapshot())
})
