import { render, waitFor } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { MockedTestProvider } from '@sourcegraph/shared/src/testing/apollo'

import { SetupWizard } from './SetupWizard'

describe('SetupWizard', () => {
    const origContext = window.context
    afterEach(() => {
        window.context = origContext
    })

    const setup = () => {
        window.context = {
            extsvcConfigFileExists: false,
            extsvcConfigAllowEdits: false,
        } as any

        return render(
            <MemoryRouter>
                <MockedTestProvider mocks={[]}>
                    <SetupWizard telemetryService={NOOP_TELEMETRY_SERVICE} />
                </MockedTestProvider>
            </MemoryRouter>
        )
    }

    it('should render correctly', async () => {
        const { getByText, asFragment } = setup()
        await waitFor(() => expect(getByText('Add remote repositories')).toBeInTheDocument())
        expect(asFragment()).toMatchSnapshot()
    })
})
