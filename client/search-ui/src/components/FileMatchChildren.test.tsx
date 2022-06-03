import { cleanup } from '@testing-library/react'
import * as H from 'history'
import _VisibilitySensor from 'react-visibility-sensor'
import { of } from 'rxjs'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import {
    RESULT,
    HIGHLIGHTED_FILE_LINES_SIMPLE_REQUEST,
    NOOP_SETTINGS_CASCADE,
    HIGHLIGHTED_FILE_LINES,
} from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { MockVisibilitySensor } from './CodeExcerpt.test'
import { FileMatchChildren } from './FileMatchChildren'

jest.mock('react-visibility-sensor', (): typeof _VisibilitySensor => ({ children, onChange }) => (
    <>
        <MockVisibilitySensor onChange={onChange}>{children}</MockVisibilitySensor>
    </>
))

const history = H.createBrowserHistory()
history.replace({ pathname: '/search' })

const onSelect = sinon.spy()

const defaultProps = {
    location: history.location,
    matches: [
        {
            preview: 'third line of code',
            line: 3,
            highlightRanges: [{ start: 7, highlightLength: 4 }],
        },
    ],
    grouped: [
        {
            matches: [{ line: 3, character: 7, highlightLength: 4 }],
            position: { line: 3, character: 7 },
            startLine: 3,
            endLine: 4,
        },
    ],
    result: RESULT,
    allMatches: true,
    subsetMatches: 10,
    fetchHighlightedFileLineRanges: HIGHLIGHTED_FILE_LINES_SIMPLE_REQUEST,
    onSelect,
    settingsCascade: NOOP_SETTINGS_CASCADE,
    telemetryService: NOOP_TELEMETRY_SERVICE,
}

describe('FileMatchChildren', () => {
    afterAll(cleanup)

    it('does not disable the highlighting timeout', () => {
        /*
            Because disabling the timeout should only ever be done in response
            to the user asking us to do so, something that we do not do for
            file matches because falling back to plaintext rendering is most
            ideal.
        */
        const fetchHighlightedFileLineRanges = sinon.spy(context => of(HIGHLIGHTED_FILE_LINES))
        renderWithBrandedContext(
            <FileMatchChildren {...defaultProps} fetchHighlightedFileLineRanges={fetchHighlightedFileLineRanges} />
        )
        sinon.assert.calledOnce(fetchHighlightedFileLineRanges)
        sinon.assert.calledWithMatch(fetchHighlightedFileLineRanges, { disableTimeout: false })
    })
})
