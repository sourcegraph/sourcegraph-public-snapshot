import { render } from '@testing-library/react'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { CommunitySearchContextsPanel } from './CommunitySearchContextPanel'

describe('CommunitySearchContextPanel', () => {
    test('renders correctly', () => {
        const props = {
            authenticatedUser: null,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        expect(render(<CommunitySearchContextsPanel {...props} />).asFragment()).toMatchSnapshot()
    })
})
