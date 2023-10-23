import { MockedProvider } from '@apollo/client/testing'
import { afterAll, beforeAll, describe, expect, it } from '@jest/globals'
import { render, type RenderResult } from '@testing-library/react'
import { spy } from 'sinon'

import { waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'

import { FuzzyWrapper, FUZZY_FILES_MOCK } from './FuzzyFinder.mocks'

describe('FuzzyModal', () => {
    const originalScrollIntoView = Element.prototype.scrollIntoView
    beforeAll(() => {
        // scrollIntoView is not supported in JSDOM, so we mock it for this one test
        // https://github.com/jsdom/jsdom/issues/1695
        Element.prototype.scrollIntoView = spy()
    })
    afterAll(() => {
        Element.prototype.scrollIntoView = originalScrollIntoView
    })

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
