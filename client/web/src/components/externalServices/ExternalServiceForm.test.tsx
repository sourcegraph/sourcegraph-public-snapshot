jest.mock('../../settings/DynamicallyImportedMonacoSettingsEditor', () => ({
    DynamicallyImportedMonacoSettingsEditor: () => 'DynamicallyImportedMonacoSettingsEditor',
}))

import * as H from 'history'
import { noop } from 'rxjs'

import { ExternalServiceKind } from '@sourcegraph/shared/src/schema'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'

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
            />
        )
        expect(component.asFragment()).toMatchSnapshot()
    })
})
