import * as H from 'history'
import { noop } from 'rxjs'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { ExternalServiceKind } from '../../graphql-operations'

import { ExternalServiceForm } from './ExternalServiceForm'

jest.mock('../../settings/DynamicallyImportedMonacoSettingsEditor', () => ({
    DynamicallyImportedMonacoSettingsEditor: () => 'DynamicallyImportedMonacoSettingsEditor',
}))

describe('ExternalServiceForm', () => {
    const baseProps = {
        history: H.createMemoryHistory(),
        isLightTheme: true,
        onSubmit: noop,
        onChange: noop,
        jsonSchema: { $id: 'json-schema-id' },
        editorActions: [],
        externalServicesFromFile: false,
        allowEditExternalServicesWithFile: false,
    }

    test('create GitHub', () => {
        const component = renderWithBrandedContext(
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
                telemetryRecorder={noOpTelemetryRecorder}
            />
        )
        expect(component.asFragment()).toMatchSnapshot()
    })
    test('edit GitHub', () => {
        const component = renderWithBrandedContext(
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
                telemetryRecorder={noOpTelemetryRecorder}
            />
        )
        expect(component.asFragment()).toMatchSnapshot()
    })
    test('edit GitHub, loading', () => {
        const component = renderWithBrandedContext(
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
                telemetryRecorder={noOpTelemetryRecorder}
            />
        )
        expect(component.asFragment()).toMatchSnapshot()
    })
})
