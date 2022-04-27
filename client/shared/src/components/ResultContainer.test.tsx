import { cleanup, fireEvent, getByTestId, getByText } from '@testing-library/react'
import * as H from 'history'
import FileIcon from 'mdi-react/FileIcon'
import sinon from 'sinon'

import { NOOP_TELEMETRY_SERVICE } from '../telemetry/telemetryService'
import { renderWithBrandedContext } from '../testing'
import {
    MULTIPLE_MATCH_RESULT,
    HIGHLIGHTED_FILE_LINES_SIMPLE_REQUEST,
    NOOP_SETTINGS_CASCADE,
} from '../testing/searchTestHelpers'

import { FileMatchChildren } from './FileMatchChildren'
import { RepoFileLink } from './RepoFileLink'
import { ResultContainer } from './ResultContainer'

describe('ResultContainer', () => {
    afterAll(cleanup)

    const history = H.createBrowserHistory()
    history.replace({ pathname: '/search' })
    const onSelect = sinon.spy()
    const expandedMatchGroups = {
        matches: [
            {
                preview: '\t"net/http/httptest"',
                line: 11,
                highlightRanges: [{ start: 15, highlightLength: 4 }],
            },
            {
                preview: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
                line: 39,
                highlightRanges: [{ start: 11, highlightLength: 4 }],
            },
            {
                preview: 'func TestTokenRequest(t *testing.T) {',
                line: 73,
                highlightRanges: [{ start: 5, highlightLength: 4 }],
            },
            {
                preview: '\tts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {',
                line: 117,
                highlightRanges: [{ start: 11, highlightLength: 4 }],
            },
            {
                preview: '\t\tio.WriteString(w, `{"access_token": "foo", "refresh_token": "bar"}`)',
                line: 134,
                highlightRanges: [{ start: 8, highlightLength: 4 }],
            },
        ],
        grouped: [
            {
                matches: [{ line: 11, character: 15, highlightLength: 4, isInContext: true }],
                position: { line: 11, character: 15 },
                startLine: 11,
                endLine: 12,
            },
            {
                matches: [{ line: 39, character: 11, highlightLength: 4, isInContext: true }],
                position: { line: 39, character: 11 },
                startLine: 39,
                endLine: 40,
            },
            {
                matches: [{ line: 73, character: 5, highlightLength: 4, isInContext: true }],
                position: { line: 73, character: 5 },
                startLine: 73,
                endLine: 74,
            },
            {
                matches: [{ line: 117, character: 11, highlightLength: 4, isInContext: true }],
                position: { line: 117, character: 11 },
                startLine: 117,
                endLine: 118,
            },
            {
                matches: [{ line: 134, character: 8, highlightLength: 4, isInContext: true }],
                position: { line: 134, character: 8 },
                startLine: 134,
                endLine: 135,
            },
        ],
    }

    const collapsedMatchGroups = {
        matches: [
            {
                preview: '\t"net/http/httptest"',
                line: 11,
                highlightRanges: [{ start: 15, highlightLength: 4 }],
            },
        ],
        grouped: [
            {
                matches: [{ line: 11, character: 15, highlightLength: 4, isInContext: true }],
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
    }

    // Props that represent a FileMatch with multiple results, totaling more than subsetMatch.
    // The FileMatch is not expanded by default, and is collapsible. These are the props passed
    // down to ResultContainers when they are used for search results.
    const defaultProps = {
        location: history.location,
        collapsible: true,
        defaultExpanded: false,
        icon: FileIcon,
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
    }

    const findReferencesProps = {
        location: history.location,
        collapsible: true,
        defaultExpanded: true,
        icon: FileIcon,
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
    }

    it('displays only one result when collapsed, which is the equivalent of subsetMatches', () => {
        const { container } = renderWithBrandedContext(<ResultContainer {...defaultProps} />)

        const expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        // 1 is the value of subsetMatches
        expect(expandedItems.length).toBe(1)
    })

    it('expands to display all results when the expand button is clicked', () => {
        const { container } = renderWithBrandedContext(<ResultContainer {...defaultProps} />)

        let expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        // 1 is the value of subsetMatches
        expect(expandedItems.length).toBe(1)

        const button = container.querySelector('[data-testid="toggle-matches-container"]')
        expect(button).toBeVisible()

        fireEvent.click(button!)

        expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        expect(expandedItems.length).toBe(5)
    })

    it('displays the expand label when collapsed', () => {
        const { container } = renderWithBrandedContext(<ResultContainer {...defaultProps} />)
        const header = getByTestId(container, 'result-container-header')
        expect(header).toBeVisible()
        expect(getByText(container, 'Show matches')).toBeVisible()
    })

    it('displays the collapse label when expanded', () => {
        const { container } = renderWithBrandedContext(<ResultContainer {...defaultProps} />)

        const button = container.querySelector('[data-testid="toggle-matches-container"]')
        expect(button).toBeVisible()

        fireEvent.click(button!)

        expect(getByText(container, 'Hide matches')).toBeVisible()
    })

    it('displays all results by default, when allExpanded is true', () => {
        const { container } = renderWithBrandedContext(<ResultContainer {...findReferencesProps} />)

        const expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        expect(expandedItems.length).toBe(5)
    })

    it('collapses to show no results when the collapse is clicked, when allExpanded is true', () => {
        const { container } = renderWithBrandedContext(<ResultContainer {...findReferencesProps} />)

        let expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        expect(expandedItems.length).toBe(5)

        const button = container.querySelector('[data-testid="toggle-matches-container"]')
        expect(button).toBeVisible()
        fireEvent.click(button!)

        expandedItems = container.querySelectorAll('[data-testid="file-match-children-item"]')
        expect(expandedItems.length).toBe(0)
    })
})
