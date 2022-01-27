import * as DOMTestingLibrary from '@testing-library/dom'
import * as DOMQueries from '@testing-library/dom/types/queries'
import * as ReactTestingLibrary from '@testing-library/react'
import userEvent from '@testing-library/user-event'

import { waitForNextApolloResponse } from '@sourcegraph/shared/src/testing/apollo'

/** We either have the render if using React, or the HTML element if using the DOM */
type HTMLContainer = HTMLElement

interface UserActionMap {
    selectBranchTab: () => Promise<void>
    selectTagTab: () => Promise<void>
    selectCommitTab: () => Promise<void>
    focusGitReference: () => void
}

export const revisionPopoverUserActions = (renderResult: HTMLContainer): UserActionMap => {
    const boundFunctions:
        | DOMTestingLibrary.BoundFunctions<typeof DOMQueries>
        | ReactTestingLibrary.BoundFunctions<typeof DOMQueries> = DOMTestingLibrary.within<typeof DOMQueries>(
        renderResult
    )

    return {
        selectBranchTab: async () => {
            const branchTab = boundFunctions.getByText('Branches')
            userEvent.click(branchTab)
            await waitForNextApolloResponse()
        },
        selectTagTab: async () => {
            const tagsTab = boundFunctions.getByText('Tags')
            userEvent.click(tagsTab)
            await waitForNextApolloResponse()
        },
        selectCommitTab: async () => {
            const commitsTab = boundFunctions.getByText('Commits')
            userEvent.click(commitsTab)
            await waitForNextApolloResponse()
        },
        focusGitReference: () => {
            const gitReference = boundFunctions.getAllByTestId('git-ref-node')[0]
            gitReference.focus()
        },
    }
}
