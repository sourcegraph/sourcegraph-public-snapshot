import { cleanup, getAllByTestId, getByTestId } from '@testing-library/react'
import FileIcon from 'mdi-react/FileIcon'
import { spy } from 'sinon'
import { afterAll, describe, expect, it } from 'vitest'

import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    HIGHLIGHTED_FILE_LINES_REQUEST,
    NOOP_SETTINGS_CASCADE,
    CHUNK_MATCH_RESULT,
} from '@sourcegraph/shared/src/testing/searchTestHelpers'

import '@sourcegraph/shared/src/testing/mockReactVisibilitySensor'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { FileContentSearchResult } from './FileContentSearchResult'

describe('FileContentSearchResult', () => {
    afterAll(cleanup)
    const defaultProps = {
        index: 0,
        result: CHUNK_MATCH_RESULT,
        icon: FileIcon,
        onSelect: spy(),
        defaultExpanded: true,
        showAllMatches: true,
        fetchHighlightedFileLineRanges: HIGHLIGHTED_FILE_LINES_REQUEST,
        settingsCascade: NOOP_SETTINGS_CASCADE,
        telemetryService: NOOP_TELEMETRY_SERVICE,
        telemetryRecorder: noOpTelemetryRecorder,
    }

    it('renders one result container', () => {
        const { container } = renderWithBrandedContext(<FileContentSearchResult {...defaultProps} />)
        expect(getByTestId(container, 'result-container')).toBeVisible()
        expect(getAllByTestId(container, 'result-container').length).toBe(1)
        expect(getAllByTestId(container, 'file-search-result').length).toBe(1)
        expect(getAllByTestId(container, 'file-match-children').length).toBe(1)
    })

    // TODO: add test that the collapse shows if there are too many results
})
