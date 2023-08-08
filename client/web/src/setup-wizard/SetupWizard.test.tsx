import { render, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { createFlagMock } from '../featureFlags/createFlagMock'

import { SetupWizard } from './SetupWizard'

describe('SetupWizard', () => {
    const origContext = window.context
    afterEach(() => {
        window.context = origContext
    })

    const setup = (setupChecklistEnabled = false) => {
        const MOCKS = [createFlagMock('setup-checklist', setupChecklistEnabled)]
        window.context = {
            extsvcConfigFileExists: false,
            extsvcConfigAllowEdits: false,
        } as any

        return render(
            <MemoryRouter>
                <MockedTestProvider mocks={MOCKS}>
                    <SetupWizard telemetryService={NOOP_TELEMETRY_SERVICE} />
                </MockedTestProvider>
            </MemoryRouter>
        )
    }

    it('should render correctly', async () => {
        const { getByText, asFragment } = setup(false)
        await waitFor(() => expect(getByText('Add remote repositories')).toBeInTheDocument())
        expect(asFragment()).toMatchSnapshot()
    })

    it('should render new checklist steps if feature flag is enabled', async () => {
        const { getByText, asFragment } = setup(true)
        // wait until new steps get rendered
        await waitFor(() => expect(getByText('Add license')).toBeInTheDocument())
        expect(asFragment()).toMatchSnapshot()
    })
})
