import React from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { CampaignChangesets } from './CampaignChangesets'
import { createRenderer } from 'react-test-renderer/shallow'
import * as H from 'history'
import { of, Subject } from 'rxjs'
import { NOOP_TELEMETRY_SERVICE } from '../../../../../../shared/src/telemetry/telemetryService'

describe('CampaignChangesets', () => {
    const history = H.createMemoryHistory()
    test('renders', () =>
        expect(
            createRenderer().render(
                <CampaignChangesets
                    queryChangesets={() =>
                        of({ nodes: [{ id: '0' } as GQL.IExternalChangeset] } as GQL.IChangesetConnection)
                    }
                    campaign={{ id: '123', closedAt: null }}
                    history={history}
                    location={history.location}
                    isLightTheme={true}
                    campaignUpdates={new Subject<void>()}
                    changesetUpdates={new Subject<void>()}
                    extensionsController={undefined as any}
                    platformContext={undefined as any}
                    telemetryService={NOOP_TELEMETRY_SERVICE}
                />
            )
        ).toMatchSnapshot())
})
