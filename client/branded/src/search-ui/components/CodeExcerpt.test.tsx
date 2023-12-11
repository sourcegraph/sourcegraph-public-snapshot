import { cleanup, getByText, render } from '@testing-library/react'
import { of } from 'rxjs'
import { map } from 'rxjs/operators'
import { afterAll, describe, expect, it } from 'vitest'

import {
    HIGHLIGHTED_FILE_LINES,
    HIGHLIGHTED_FILE_LINES_LONG,
    HIGHLIGHTED_FILE_LINES_SIMPLE,
    FILE_LINES_SIMPLE,
} from '@sourcegraph/shared/src/testing/searchTestHelpers'

import '@sourcegraph/shared/src/testing/mockReactVisibilitySensor'

import { CodeExcerpt } from './CodeExcerpt'

describe('CodeExcerpt', () => {
    afterAll(cleanup)

    const startLine = 0
    const endLine = 3
    const defaultProps = {
        blobLines: FILE_LINES_SIMPLE,
        repoName: 'github.com/golang/oauth2',
        commitID: 'e64efc72b421e893cbf63f17ba2221e7d6d0b0f3',
        filePath: '.travis.yml',
        highlightRanges: [
            { startLine: 0, startCharacter: 6, endLine: 0, endCharacter: 10 },
            { startLine: 1, startCharacter: 7, endLine: 1, endCharacter: 11 },
            { startLine: 2, startCharacter: 6, endLine: 2, endCharacter: 10 },
        ],
        startLine,
        endLine,
        isLightTheme: false,
        className: 'file-match__item-code-excerpt',
        fetchHighlightedFileRangeLines: () =>
            of(HIGHLIGHTED_FILE_LINES_SIMPLE).pipe(map(ranges => ranges[0].slice(startLine, endLine))),
    }

    it('renders correct number of rows', () => {
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        expect(container.querySelectorAll('[data-testid="code-excerpt"] tr').length).toBe(3)
    })

    it('renders the line number container on each row', () => {
        // We can't evaluate the content of these containers since the numbers are drawn
        // by CSS. This is a proxy to make sure the <td> tags with data-line attributes
        // at least exist.
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        const dataLines = container.querySelectorAll('[data-line]')
        expect(dataLines).toHaveLength(3)
    })

    it('renders the code portion of each row', () => {
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        expect(getByText(container, 'first of code')).toBeVisible()
        expect(getByText(container, 'second of code')).toBeVisible()
        expect(getByText(container, 'third of code')).toBeVisible()
    })

    it('highlights matches correctly', () => {
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        const highlightedSpans = container.querySelectorAll('.match-highlight')
        expect(highlightedSpans.length).toBe(3)
        for (const span of highlightedSpans) {
            expect(span.textContent === 'line')
        }
    })

    it('highlights matches correctly in syntax highlighted code blocks', () => {
        const highlightRanges = [{ startLine: 12, startCharacter: 7, endLine: 12, endCharacter: 11 }]
        const startLine = 0
        const endLine = 13
        const { container } = render(
            <CodeExcerpt
                {...defaultProps}
                startLine={startLine}
                endLine={endLine}
                highlightRanges={highlightRanges}
                fetchHighlightedFileRangeLines={() =>
                    of(HIGHLIGHTED_FILE_LINES).pipe(map(ranges => ranges[0].slice(startLine, endLine)))
                }
            />
        )
        const highlightedSpans = container.querySelectorAll('.match-highlight')
        expect(highlightedSpans.length).toBe(1)
        for (const span of highlightedSpans) {
            expect(span.textContent === 'test')
        }
    })

    it('displays the correct number of lines, with no highlight matches on the context lines', () => {
        const highlightRanges = [
            { startLine: 0, startCharacter: 0, endLine: 0, endCharacter: 1 },
            { startLine: 1, startCharacter: 0, endLine: 1, endCharacter: 1 },
            { startLine: 2, startCharacter: 0, endLine: 2, endCharacter: 1 },
            { startLine: 4, startCharacter: 0, endLine: 4, endCharacter: 1 },
            { startLine: 7, startCharacter: 0, endLine: 7, endCharacter: 1 },
            { startLine: 10, startCharacter: 0, endLine: 10, endCharacter: 1 },
            { startLine: 11, startCharacter: 0, endLine: 11, endCharacter: 1 },
            { startLine: 16, startCharacter: 0, endLine: 16, endCharacter: 1 },
            { startLine: 17, startCharacter: 0, endLine: 17, endCharacter: 1 },
            { startLine: 19, startCharacter: 0, endLine: 19, endCharacter: 1 },
        ]
        const startLine = 0
        const endLine = 22
        const { container } = render(
            <CodeExcerpt
                {...defaultProps}
                startLine={startLine}
                endLine={endLine}
                highlightRanges={highlightRanges}
                fetchHighlightedFileRangeLines={() =>
                    of(HIGHLIGHTED_FILE_LINES_LONG).pipe(map(ranges => ranges[0].slice(startLine, endLine)))
                }
            />
        )
        const rows = container.querySelectorAll('tr')
        expect(rows.length).toBe(22)
    })

    it('displays the correct number of lines, with matches on the context line', () => {
        const highlightRanges = [
            { startLine: 0, startCharacter: 0, endLine: 0, endCharacter: 1 },
            { startLine: 1, startCharacter: 0, endLine: 1, endCharacter: 1 },
            { startLine: 2, startCharacter: 0, endLine: 2, endCharacter: 1 },
            { startLine: 4, startCharacter: 0, endLine: 4, endCharacter: 1 },
            { startLine: 7, startCharacter: 0, endLine: 7, endCharacter: 1 },
            { startLine: 10, startCharacter: 0, endLine: 10, endCharacter: 1 },
            { startLine: 11, startCharacter: 0, endLine: 11, endCharacter: 1 },
            { startLine: 16, startCharacter: 0, endLine: 16, endCharacter: 1 },
            { startLine: 17, startCharacter: 0, endLine: 17, endCharacter: 1 },
            { startLine: 19, startCharacter: 0, endLine: 19, endCharacter: 1 },
            { startLine: 20, startCharacter: 0, endLine: 20, endCharacter: 1 },
        ]
        const startLine = 0
        const endLine = 22
        const { container } = render(
            <CodeExcerpt
                {...defaultProps}
                startLine={startLine}
                endLine={endLine}
                highlightRanges={highlightRanges}
                fetchHighlightedFileRangeLines={() =>
                    of(HIGHLIGHTED_FILE_LINES_LONG).pipe(map(ranges => ranges[0].slice(startLine, endLine)))
                }
            />
        )
        const rows = container.querySelectorAll('tr')
        expect(rows.length).toBe(22)
    })
})
