import { cleanup, getAllByTestId, getByTestId } from '@testing-library/react'
import { createBrowserHistory } from 'history'
import FileIcon from 'mdi-react/FileIcon'
import _VisibilitySensor from 'react-visibility-sensor'
import sinon from 'sinon'

import { MatchGroup } from '@sourcegraph/shared/src/components/ranking/PerFileResultRanking'
import { ContentMatch } from '@sourcegraph/shared/src/search/stream'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { renderWithBrandedContext } from '@sourcegraph/shared/src/testing'
import {
    HIGHLIGHTED_FILE_LINES_REQUEST,
    NOOP_SETTINGS_CASCADE,
    RESULT,
} from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { MockVisibilitySensor } from './CodeExcerpt.test'
import { FileSearchResult, limitGroup } from './FileSearchResult'

jest.mock('react-visibility-sensor', (): typeof _VisibilitySensor => ({ children, onChange }) => (
    <>
        <MockVisibilitySensor onChange={onChange}>{children}</MockVisibilitySensor>
    </>
))

describe('FileSearchResult', () => {
    afterAll(cleanup)
    const history = createBrowserHistory()
    history.replace({ pathname: '/search' })
    const defaultProps = {
        location: history.location,
        result: RESULT,
        icon: FileIcon,
        onSelect: sinon.spy(),
        expanded: true,
        showAllMatches: true,
        fetchHighlightedFileLineRanges: HIGHLIGHTED_FILE_LINES_REQUEST,
        settingsCascade: NOOP_SETTINGS_CASCADE,
        telemetryService: NOOP_TELEMETRY_SERVICE,
    }

    it('renders one result container', () => {
        const { container } = renderWithBrandedContext(<FileSearchResult {...defaultProps} />)
        expect(getByTestId(container, 'result-container')).toBeVisible()
        expect(getAllByTestId(container, 'result-container').length).toBe(1)
    })

    it('correctly shows number of context lines when search.contextLines setting is set', () => {
        const result: ContentMatch = {
            type: 'content',
            path: '.travis.yml',
            repository: 'github.com/golang/oauth2',
            lineMatches: [
                {
                    line: '  - go test -v golang.org/x/oauth2/...',
                    lineNumber: 4,
                    offsetAndLengths: [[7, 4]],
                },
            ],
        }
        const settingsCascade = {
            final: { 'search.contextLines': 3 },
            subjects: [
                {
                    lastID: 1,
                    settings: { 'search.contextLines': '3' },
                    extensions: null,
                    subject: {
                        __typename: 'User' as const,
                        username: 'f',
                        id: 'abc',
                        settingsURL: '/users/f/settings',
                        viewerCanAdminister: true,
                        displayName: 'f',
                    },
                },
            ],
        }
        const { container } = renderWithBrandedContext(
            <FileSearchResult {...defaultProps} result={result} settingsCascade={settingsCascade} />
        )
        const tableRows = container.querySelectorAll('[data-testid="code-excerpt"] tr')
        expect(tableRows.length).toBe(7)
    })
})

describe('limitGroup', () => {
    it('truncates a group', () => {
        const group: MatchGroup = {
            blobLines: ['line0', 'line1', 'line2'],
            matches: [
                {
                    line: 0,
                    character: 0,
                    highlightLength: 1,
                },
                {
                    line: 2,
                    character: 0,
                    highlightLength: 1,
                },
            ],
            position: {
                line: 1,
                character: 1,
            },
            startLine: 0,
            endLine: 3,
        }

        const expected: MatchGroup = {
            blobLines: ['line0', 'line1'],
            matches: [
                {
                    line: 0,
                    character: 0,
                    highlightLength: 1,
                },
            ],
            position: {
                line: 1,
                character: 1,
            },
            startLine: 0,
            endLine: 2,
        }
        const limitedGroup = limitGroup(group, 1)
        expect(limitedGroup).toStrictEqual(expected)
    })

    it('preserves a group that does not need limited', () => {
        const group: MatchGroup = {
            blobLines: ['line0', 'line1', 'line2'],
            matches: [
                {
                    line: 0,
                    character: 0,
                    highlightLength: 1,
                },
            ],
            position: {
                line: 1,
                character: 1,
            },
            startLine: 0,
            endLine: 3,
        }
        const limitedGroup = limitGroup(group, 10)
        expect(limitedGroup).toStrictEqual(group)
    })

    it('truncates a group, but saves a match if it is in the context line', () => {
        const group: MatchGroup = {
            blobLines: ['line0', 'line1', 'line2'],
            matches: [
                {
                    line: 0,
                    character: 0,
                    highlightLength: 1,
                },
                {
                    line: 1,
                    character: 0,
                    highlightLength: 1,
                },
            ],
            position: {
                line: 1,
                character: 1,
            },
            startLine: 0,
            endLine: 3,
        }

        const expected: MatchGroup = {
            blobLines: ['line0', 'line1'],
            matches: [
                {
                    line: 0,
                    character: 0,
                    highlightLength: 1,
                },
                {
                    line: 1,
                    character: 0,
                    highlightLength: 1,
                },
            ],
            position: {
                line: 1,
                character: 1,
            },
            startLine: 0,
            endLine: 2,
        }
        const limitedGroup = limitGroup(group, 1)
        expect(limitedGroup).toStrictEqual(expected)
    })

    it('truncates a group, but saves a match if it is on the last line', () => {
        const group: MatchGroup = {
            blobLines: ['line0', 'line1', 'line2'],
            matches: [
                {
                    line: 0,
                    character: 0,
                    highlightLength: 1,
                },
                {
                    line: 0,
                    character: 2,
                    highlightLength: 3,
                },
            ],
            position: {
                line: 1,
                character: 1,
            },
            startLine: 0,
            endLine: 3,
        }

        const expected: MatchGroup = {
            blobLines: ['line0', 'line1'],
            matches: [
                {
                    line: 0,
                    character: 0,
                    highlightLength: 1,
                },
                {
                    line: 0,
                    character: 2,
                    highlightLength: 3,
                },
            ],
            position: {
                line: 1,
                character: 1,
            },
            startLine: 0,
            endLine: 2,
        }
        const limitedGroup = limitGroup(group, 1)
        expect(limitedGroup).toStrictEqual(expected)
    })
})
