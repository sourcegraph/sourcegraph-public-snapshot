jest.mock('../../settings/DynamicallyImportedMonacoSettingsEditor', () => ({
    DynamicallyImportedMonacoSettingsEditor: () => 'DynamicallyImportedMonacoSettingsEditor',
}))

import { render } from '@testing-library/react'
import * as H from 'history'
import React from 'react'
import { noop } from 'rxjs'

import { ExternalServiceKind } from '@sourcegraph/shared/src/schema'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'

import { ExternalServiceForm } from './ExternalServiceForm'

describe('ExternalServiceForm', () => {
    const baseProps = {
        history: H.createMemoryHistory(),
        isLightTheme: true,
        onSubmit: noop,
        onChange: noop,
        jsonSchema: { $id: 'json-schema-id' },
        editorActions: [],
    }

    test('create GitHub', () => {
        const component = render(
            <ExternalServiceForm
                {...baseProps}
                input={{
                    kind: ExternalServiceKind.GITHUB,
                    displayName: 'GitHub',
                    config: '{}',
                }}
                mode="create"
                loading={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )
        expect(component.asFragment()).toMatchSnapshot()
    })
    test('edit GitHub', () => {
        const component = render(
            <ExternalServiceForm
                {...baseProps}
                input={{
                    kind: ExternalServiceKind.GITHUB,
                    displayName: 'GitHub',
                    config: '{}',
                }}
                mode="create"
                loading={false}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )
        expect(component.asFragment()).toMatchSnapshot()
    })
    test('edit GitHub, loading', () => {
        const component = render(
            <ExternalServiceForm
                {...baseProps}
                input={{
                    kind: ExternalServiceKind.GITHUB,
                    displayName: 'GitHub',
                    config: '{}',
                }}
                mode="create"
                loading={true}
                telemetryService={NOOP_TELEMETRY_SERVICE}
            />
        )
        expect(component.asFragment()).toMatchSnapshot()
    })
})
