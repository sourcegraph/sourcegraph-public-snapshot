import React from 'react'
import { CampaignSpecTab } from './CampaignSpecTab'
import { mount } from 'enzyme'
import { CampaignFields } from '../../../graphql-operations'

jest.mock('mdi-react/FileDownloadIcon', () => 'FileDownloadIcon')
jest.mock('../../../../../shared/src/util/markdown', () => ({ highlightCodeSafe: () => 'code' }))

const ALICE: CampaignFields['initialApplier'] | CampaignFields['lastApplier'] = {
    username: 'alice',
    url: 'https://example.com/alice',
}

describe('CampaignSpecTab', () => {
    test('last apply was the (initial) creation', () => {
        expect(
            mount(
                <CampaignSpecTab
                    campaign={{
                        name: 'c',
                        createdAt: '2020-01-01T15:00:00Z',
                        lastAppliedAt: '2020-01-01T15:00:00Z',
                        lastApplier: ALICE,
                    }}
                    originalInput="x"
                />
            )
        ).toMatchSnapshot()
    })

    test('last apply was an update (after creation)', () => {
        expect(
            mount(
                <CampaignSpecTab
                    campaign={{
                        name: 'c',
                        createdAt: '2020-01-01T15:00:00Z',
                        lastAppliedAt: '2020-02-03T16:07:08Z',
                        lastApplier: ALICE,
                    }}
                    originalInput="x"
                />
            )
        ).toMatchSnapshot()
    })
})
