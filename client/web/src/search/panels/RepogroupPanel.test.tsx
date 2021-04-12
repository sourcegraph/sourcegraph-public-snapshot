import { mount } from 'enzyme'
import React from 'react'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { RepogroupPanel } from './RepogroupPanel'

describe('RepogroupPanel', () => {
    test('renders correctly', () => {
        const props = {
            authenticatedUser: null,
            telemetryService: NOOP_TELEMETRY_SERVICE,
        }

        expect(mount(<RepogroupPanel {...props} />)).toMatchSnapshot()
    })
})
