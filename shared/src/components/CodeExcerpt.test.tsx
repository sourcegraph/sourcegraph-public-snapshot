import _VisibilitySensor from 'react-visibility-sensor'

jest.mock(
    'react-visibility-sensor',
    (): typeof _VisibilitySensor => ({ children, onChange }) => {
        if (onChange) {
            onChange(true)
        }
        return <>{children}</>
    }
)

import * as React from 'react'
import { cleanup, getAllByText, getByText, render } from 'react-testing-library'
import { HIGHLIGHTED_FILE_LINES_SIMPLE_REQUEST } from '../../../web/src/search/results/testHelpers'
import { CodeExcerpt } from './CodeExcerpt'

// TODO(attfarhan): Factor out isVisible flag into a prop in CodeExcerpt so we can test the code excerpt
// after it has fetched highlighted file lines.
describe('CodeExcerpt', () => {
    afterAll(cleanup)

    const defaultProps = {
        repoName: 'github.com/golang/oauth2',
        commitID: 'e64efc72b421e893cbf63f17ba2221e7d6d0b0f3',
        filePath: '.travis.yml',
        highlightRanges: [
            { line: 0, character: 6, highlightLength: 4 },
            { line: 1, character: 7, highlightLength: 4 },
            { line: 2, character: 6, highlightLength: 4 },
        ],
        isLightTheme: false,
        className: 'file-match__item-code-excerpt',
        fetchHighlightedFileLines: HIGHLIGHTED_FILE_LINES_SIMPLE_REQUEST,
    }

    it('renders all code lines', () => {
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        expect(container.querySelectorAll('.code-excerpt tr').length).toBe(3)
    })

    it('displays line numbers', () => {
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        expect(getAllByText(container, '1').length).toBe(1)
        expect(getAllByText(container, '2').length).toBe(1)
        expect(getAllByText(container, '3').length).toBe(1)
    })

    it('renders all code lines', () => {
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        expect(getByText(container, 'first of code')).toBeTruthy()
        expect(getByText(container, 'second of code')).toBeTruthy()
        expect(getByText(container, 'third of code')).toBeTruthy()
        expect(getAllByText(container, 'line').length).toBe(3)
    })

    it('highlights matches correctly', () => {
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        const highlightedSpans = container.querySelectorAll('.selection-highlight')
        expect(highlightedSpans.length).toBe(3)
        for (const span of highlightedSpans) {
            expect(span.textContent === 'line')
        }
    })
})
