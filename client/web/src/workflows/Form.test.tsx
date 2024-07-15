import { describe, test } from 'vitest'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { WorkflowForm } from './Form'

describe('WorkflowForm', () => {
    test('renders', () => {
        renderWithBrandedContext(
            <WorkflowForm
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
