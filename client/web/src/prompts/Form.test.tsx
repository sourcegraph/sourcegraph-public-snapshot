import { describe, test } from 'vitest'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { PromptForm } from './Form'

describe('PromptForm', () => {
    test('renders', () => {
        renderWithBrandedContext(
            <PromptForm
                submitLabel="Submit"
                initialValue={{}}
                namespaceField={null}
                onSubmit={() => {}}
                loading={false}
                error={null}
                telemetryRecorder={noOpTelemetryRecorder}
            />
        )
    })
})
