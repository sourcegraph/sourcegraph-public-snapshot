import React from 'react'
import { CampaignDetails } from './CampaignDetails'
import * as H from 'history'
import { of } from 'rxjs'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../shared/src/telemetry/telemetryService'
import { PageTitle } from '../../../components/PageTitle'
import { registerHighlightContributions } from '../../../../../shared/src/highlight/contributions'
import { mount } from 'enzyme'

// This is idempotent, so calling it in multiple tests is not a problem.
registerHighlightContributions()

const history = H.createMemoryHistory()

describe('CampaignDetails', () => {
    afterEach(() => {
        PageTitle.titleSet = false
    })

    const renderCampaignDetails = ({ viewerCanAdminister }: { viewerCanAdminister: boolean }) => (
        <CampaignDetails
            campaignID="c"
            history={history}
            location={history.location}
            isLightTheme={true}
            extensionsController={undefined as any}
            platformContext={undefined as any}
            telemetryService={NOOP_TELEMETRY_SERVICE}
            _fetchCampaignById={() =>
                of({
                    __typename: 'Campaign',
                    id: 'c',
                    name: 'n',
                    description: 'd',
                    initialApplier: { username: 'alice', avatarURL: 'http://test.test/avatar' },
                    changesets: { totalCount: 0, stats: { total: 10, closed: 0, merged: 0, open: 8, unpublished: 2 } },
                    changesetCountsOverTime: [],
                    viewerCanAdminister,
                    branch: 'awesome-branch',
                    createdAt: '2020-01-01',
                    updatedAt: '2020-01-01',
                    closedAt: null,
                    diffStat: {
                        added: 5,
                        changed: 3,
                        deleted: 2,
                    },
                    namespace: {
                        namespaceName: 'alice',
                    },
                })
            }
        />
    )

    for (const viewerCanAdminister of [true, false]) {
        describe(`viewerCanAdminister: ${String(viewerCanAdminister)}`, () => {
            test('viewing existing', () => {
                const rendered = mount(renderCampaignDetails({ viewerCanAdminister }))
                expect(rendered).toMatchSnapshot()
            })
        })
    }
})
