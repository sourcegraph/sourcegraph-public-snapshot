import * as React from 'react'

import { cleanup, getByText, render } from '@testing-library/react'
import _VisibilitySensor from 'react-visibility-sensor'
import { of } from 'rxjs'
import { map } from 'rxjs/operators'

import {
    HIGHLIGHTED_FILE_LINES,
    HIGHLIGHTED_FILE_LINES_LONG,
    HIGHLIGHTED_FILE_LINES_SIMPLE,
} from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { CodeExcerpt } from './CodeExcerpt'

export class MockVisibilitySensor extends React.Component<{ onChange?: (isVisible: boolean) => void }> {
    constructor(props: { onChange?: (isVisible: boolean) => void }) {
        super(props)
        if (props.onChange) {
            props.onChange(true)
        }
    }

    public render(): JSX.Element {
        return <>{this.props.children}</>
    }
}

jest.mock('react-visibility-sensor', (): typeof _VisibilitySensor => ({ children, onChange }) => (
    <>
        <MockVisibilitySensor onChange={onChange}>{children}</MockVisibilitySensor>
    </>
))

describe('CodeExcerpt', () => {
    afterAll(cleanup)

    const startLine = 0
    const endLine = 3
    const defaultProps = {
        repoName: 'github.com/golang/oauth2',
        commitID: 'e64efc72b421e893cbf63f17ba2221e7d6d0b0f3',
        filePath: '.travis.yml',
        highlightRanges: [
            { line: 0, character: 6, highlightLength: 4 },
            { line: 1, character: 7, highlightLength: 4 },
            { line: 2, character: 6, highlightLength: 4 },
        ],
        startLine,
        endLine,
        isLightTheme: false,
        className: 'file-match__item-code-excerpt',
        fetchHighlightedFileRangeLines: () =>
            of(HIGHLIGHTED_FILE_LINES_SIMPLE).pipe(map(ranges => ranges[0].slice(startLine, endLine))),
        isFirst: false,
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
        expect(dataLines.length).toMatchInlineSnapshot('3')
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
        const highlightRanges = [{ line: 12, character: 7, highlightLength: 4 }]
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
            { line: 0, character: 0, highlightLength: 1 },
            { line: 1, character: 0, highlightLength: 1 },
            { line: 2, character: 0, highlightLength: 1 },
            { line: 4, character: 0, highlightLength: 1 },
            { line: 7, character: 0, highlightLength: 1 },
            { line: 10, character: 0, highlightLength: 1 },
            { line: 11, character: 0, highlightLength: 1 },
            { line: 16, character: 0, highlightLength: 1 },
            { line: 17, character: 0, highlightLength: 1 },
            { line: 19, character: 0, highlightLength: 1 },
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
            { line: 0, character: 0, highlightLength: 1 },
            { line: 1, character: 0, highlightLength: 1 },
            { line: 2, character: 0, highlightLength: 1 },
            { line: 4, character: 0, highlightLength: 1 },
            { line: 7, character: 0, highlightLength: 1 },
            { line: 10, character: 0, highlightLength: 1 },
            { line: 11, character: 0, highlightLength: 1 },
            { line: 16, character: 0, highlightLength: 1 },
            { line: 17, character: 0, highlightLength: 1 },
            { line: 19, character: 0, highlightLength: 1 },
            { line: 20, character: 0, highlightLength: 1 },
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
