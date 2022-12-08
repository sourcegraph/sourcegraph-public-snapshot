import { MockedProvider } from '@apollo/client/testing'
import { render, RenderResult } from '@testing-library/react'

import { waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'

import { FuzzyWrapper, FUZZY_FILES_MOCK } from './FuzzyFinder.mocks'

describe('FuzzyModal', () => {
    it('displays all, repos, symbols and files tabs with default experimentalFeatures', async () => {
        const result: RenderResult = render(
            <MockedProvider mocks={[FUZZY_FILES_MOCK]}>
                <FuzzyWrapper
                    url="/github.com/sourcegraph/sourcegraph@main"
                    experimentalFeatures={{}}
                    initialQuery="clientb"
                />
            </MockedProvider>
        )
        await waitForNextApolloResponse()

        for (const tab of ['all', 'repos', 'symbols', 'files']) {
            expect(result.queryByTestId(tab)).toBeInTheDocument()
        }

        for (const tab of ['actions', 'lines']) {
            expect(result.queryByTestId(tab)).not.toBeInTheDocument()
        }
    })
})
