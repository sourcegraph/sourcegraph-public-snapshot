import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

import { CommunitySearchContextsPanel } from './CommunitySearchContextPanel'

describe('CommunitySearchContextPanel', () => {
    test('renders correctly', () => {
        const props = {
            authenticatedUser: null,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        expect(renderWithBrandedContext(<CommunitySearchContextsPanel {...props} />).asFragment()).toMatchSnapshot()
    })
})
