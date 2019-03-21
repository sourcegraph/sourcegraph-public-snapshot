import * as React from 'react'
import { cleanup, getAllByText, render } from 'react-testing-library'
import { HIGHLIGHTED_FILE_LINES_REQUEST } from '../../../web/src/search/results/testHelpers'
import { CodeExcerpt } from './CodeExcerpt'

// TODO(attfarhan): Factor out isVisible flag into a prop in CodeExcerpt so we can test the code excerpt
// after it has fetched highlighted file lines.
describe('CodeExcerpt', () => {
    afterAll(cleanup)

    const defaultProps = {
        repoName: 'github.com/golang/oauth2',
        commitID: 'e64efc72b421e893cbf63f17ba2221e7d6d0b0f3',
        filePath: '.travis.yml',
        highlightRanges: [{ line: 12, character: 7, highlightLength: 4 }],
        isLightTheme: false,
        className: 'file-match__item-code-excerpt',
        fetchHighlightedFileLines: HIGHLIGHTED_FILE_LINES_REQUEST,
    }

    it('renders the code excerpt', () => {
        render(<CodeExcerpt {...defaultProps} />)
    })

    it('renders both code lines', () => {
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        expect(container.querySelectorAll('.code-excerpt tr').length).toBe(2)
    })

    it('displays line numbers', () => {
        const { container } = render(<CodeExcerpt {...defaultProps} />)
        expect(getAllByText(container, '12').length).toBe(1)
        expect(getAllByText(container, '13').length).toBe(1)
    })
})
