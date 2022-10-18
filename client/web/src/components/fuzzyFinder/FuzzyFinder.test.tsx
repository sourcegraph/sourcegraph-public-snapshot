import { MockedProvider } from '@apollo/client/testing'
import { render, RenderResult } from '@testing-library/react'

import { waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'

import { FuzzyWrapper, FUZZY_FILES_MOCK } from './FuzzyFinder.mocks'

describe('FuzzyModal', () => {
    it('displays only files with default experimentalFeatures default', async () => {
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
        expect(result.getByTestId('fuzzy-modal-summary')).toHaveTextContent('12 results - 14 totals')
        expect(result.getByTestId('fuzzy-modal-header')).toHaveTextContent('Find files')
    })
})
