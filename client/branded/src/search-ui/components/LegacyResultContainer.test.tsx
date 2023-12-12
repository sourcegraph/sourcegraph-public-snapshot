import { cleanup, fireEvent, getByTestId, getByText } from '@testing-library/react'
import * as H from 'history'
import FileIcon from 'mdi-react/FileIcon'
import sinon from 'sinon'

import { noOpTelemetryRecorder } from '@sourcegraph/shared/src/telemetry'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    MULTIPLE_MATCH_RESULT,
    HIGHLIGHTED_FILE_LINES_SIMPLE_REQUEST,
    NOOP_SETTINGS_CASCADE,
} from '@sourcegraph/shared/src/testing/searchTestHelpers'
import { renderWithBrandedContext } from '@sourcegraph/wildcard/src/testing'

import { FileMatchChildren } from './FileMatchChildren'
import { LegacyResultContainer } from './LegacyResultContainer'
import { RepoFileLink } from './RepoFileLink'

describe('LegacyResultContainer', () => {
    afterAll(cleanup)

    const history = H.createBrowserHistory()
    history.replace({ pathname: '/search' })
    const onSelect = sinon.spy()
    const expandedMatchGroups = {
        matches: [
            {
                content: '\t"net/http/httptest"',
                startLine: 11,
                highlightRanges: [{ startLine: 11, startCharacter: 15, endLine: 11, endCharacter: 19 }],
            },
            {
                content: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
                startLine: 39,
                highlightRanges: [{ startLine: 39, startCharacter: 11, endLine: 39, endCharacter: 15 }],
            },
            {
                content: 'func TestTokenRequest(t *testing.T) {',
                startLine: 73,
                highlightRanges: [{ startLine: 39, startCharacter: 5, endLine: 39, endCharacter: 9 }],
            },
            {
                content: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
                startLine: 117,
                highlightRanges: [{ startLine: 117, startCharacter: 11, endLine: 117, endCharacter: 15 }],
            },
            {
                content: '\t\tio.WriteString(w, `{"access_token": "foo", "refresh_token": "bar"}`)',
                startLine: 134,
                highlightRanges: [{ startLine: 134, startCharacter: 8, endLine: 124, endCharacter: 12 }],
            },
        ],
        grouped: [
            {
                matches: [{ startLine: 11, startCharacter: 15, endLine: 11, endCharacter: 19 }],
                position: { line: 11, character: 15 },
                startLine: 11,
                endLine: 12,
            },
            {
                matches: [{ startLine: 39, startCharacter: 11, endLine: 39, endCharacter: 15 }],
                position: { line: 39, character: 11 },
                startLine: 39,
                endLine: 40,
            },
            {
                matches: [{ startLine: 73, startCharacter: 5, endLine: 73, endCharacter: 9 }],
                position: { line: 73, character: 5 },
                startLine: 73,
                endLine: 74,
            },
            {
                matches: [{ startLine: 117, startCharacter: 11, endLine: 117, endCharacter: 15 }],
                position: { line: 117, character: 11 },
                startLine: 117,
                endLine: 118,
            },
            {
                matches: [{ startLine: 134, startCharacter: 8, endLine: 124, endCharacter: 12 }],
                position: { line: 134, character: 8 },
                startLine: 134,
                endLine: 135,
            },
        ],
    }

    const collapsedMatchGroups = {
        matches: [
            {
                content: '\t"net/http/httptest"',
                startLine: 11,
                highlightRanges: [{ startLine: 11, startCharacter: 15, endLine: 11, endCharacter: 19 }],
            },
        ],
        grouped: [
            {
                matches: [{ startLine: 11, startCharacter: 15, endLine: 11, endCharacter: 19 }],
                position: { line: 11, character: 15 },
                startLine: 11,
                endLine: 12,
            },
        ],
    }

    const fileMatchChildrenProps = {
        location: history.location,
        result: MULTIPLE_MATCH_RESULT,
        fetchHighlightedFileLineRanges: HIGHLIGHTED_FILE_LINES_SIMPLE_REQUEST,
        onSelect,
        settingsCascade: NOOP_SETTINGS_CASCADE,
        telemetryService: NOOP_TELEMETRY_SERVICE,
        telemetryRecorder: noOpTelemetryRecorder,
    }

    // Props that represent a FileMatch with multiple results, totaling more than subsetMatch.
    // The FileMatch is not expanded by default, and is collapsible. These are the props passed
    // down to ResultContainers when they are used for search results.
    const defaultProps = {
        index: 0,
        location: history.location,
        collapsible: true,
        defaultExpanded: false,
        icon: FileIcon,
        repoName: 'example.com/my/repo',
        title: (
            <RepoFileLink
                repoName="example.com/my/repo"
                repoURL="https://example.com"
                filePath="file.go"
                fileURL="https://example.com/file"
            />
        ),
        expandedChildren: <FileMatchChildren {...fileMatchChildrenProps} {...expandedMatchGroups} />,
        collapsedChildren: <FileMatchChildren {...fileMatchChildrenProps} {...collapsedMatchGroups} />,
        collapseLabel: 'Hide matches',
        expandLabel: 'Show matches',
        allExpanded: false,
        telemetryService: NOOP_TELEMETRY_SERVICE,
        telemetryRecorder: noOpTelemetryRecorder,
    }

    const findReferencesProps = {
        index: 0,
        location: history.location,
        collapsible: true,
        defaultExpanded: true,
        icon: FileIcon,
        repoName: 'example.com/my/repo',
        title: (
            <RepoFileLink
                repoName="example.com/my/repo"
                repoURL="https://example.com"
                filePath="file.go"
                fileURL="https://example.com/file"
            />
        ),
        expandedChildren: <FileMatchChildren {...fileMatchChildrenProps} {...expandedMatchGroups} />,
        telemetryService: NOOP_TELEMETRY_SERVICE,
        telemetryRecorder: noOpTelemetryRecorder,
    }

    it('displays only one result when collapsed, which is the equivalent of subsetMatches', () => {
        const { container } = renderWithBrandedContext(<LegacyResultContainer {...defaultProps} />)

        const expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        // 1 is the value of subsetMatches
        expect(expandedItems.length).toBe(1)
    })

    it('expands to display all results when the expand button is clicked', () => {
        const { container } = renderWithBrandedContext(<LegacyResultContainer {...defaultProps} />)

        let expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        // 1 is the value of subsetMatches
        expect(expandedItems.length).toBe(1)

        const button = container.querySelector('[data-testid="toggle-matches-container"]') as Element
        expect(button).toBeVisible()

        fireEvent.click(button)

        expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        expect(expandedItems.length).toBe(5)
    })

    it('displays the expand label when collapsed', () => {
        const { container } = renderWithBrandedContext(<LegacyResultContainer {...defaultProps} />)
        const header = getByTestId(container, 'result-container-header')
        expect(header).toBeVisible()
        expect(getByText(container, 'Show matches')).toBeVisible()
    })

    it('displays the collapse label when expanded', () => {
        const { container } = renderWithBrandedContext(<LegacyResultContainer {...defaultProps} />)

        const button = container.querySelector('[data-testid="toggle-matches-container"]') as Element
        expect(button).toBeVisible()

        fireEvent.click(button)

        expect(getByText(container, 'Hide matches')).toBeVisible()
    })

    it('displays all results by default, when allExpanded is true', () => {
        const { container } = renderWithBrandedContext(<LegacyResultContainer {...findReferencesProps} />)

        const expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        expect(expandedItems.length).toBe(5)
    })

    it('collapses to show no results when the collapse is clicked, when allExpanded is true', () => {
        const { container } = renderWithBrandedContext(<LegacyResultContainer {...findReferencesProps} />)

        let expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        expect(expandedItems.length).toBe(5)

        const button = container.querySelector('[data-testid="toggle-matches-container"]') as Element
        expect(button).toBeVisible()
        fireEvent.click(button)

        expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        expect(expandedItems.length).toBe(0)
    })
})
