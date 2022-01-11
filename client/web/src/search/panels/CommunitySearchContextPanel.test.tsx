import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithRouter } from '@sourcegraph/shared/src/testing/render-with-router'

import { CommunitySearchContextsPanel } from './CommunitySearchContextPanel'

describe('CommunitySearchContextPanel', () => {
    test('renders correctly', () => {
        const props = {
            authenticatedUser: null,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        expect(renderWithRouter(<CommunitySearchContextsPanel {...props} />).asFragment()).toMatchSnapshot()
    })
})
