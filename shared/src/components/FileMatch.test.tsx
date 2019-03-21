import { createBrowserHistory } from 'history'
import FileIcon from 'mdi-react/FileIcon'
import * as React from 'react'
import { cleanup, getAllByTestId, getByTestId, getByText, render } from 'react-testing-library'
import sinon from 'sinon'
import { NOOP_SETTINGS_CASCADE, RESULT, SEARCH_REQUEST } from '../../../web/src/search/results/testHelpers'
import { FileMatch, IFileMatch } from './FileMatch'
import { setLinkComponent } from './Link'

describe('FileMatch', () => {
    setLinkComponent((props: any) => <a {...props} />)

    afterAll(() => {
        setLinkComponent(null as any)
        cleanup()
    })
    const history = createBrowserHistory()
    const defaultProps = {
        location: history.location,
        result: RESULT as IFileMatch,
        icon: FileIcon,
        onSelect: sinon.spy(),
        expanded: true,
        showAllMatches: true,
        isLightTheme: true,
        fetchHighlightedFileLines: SEARCH_REQUEST,
        settingsCascade: NOOP_SETTINGS_CASCADE,
    }

    it('renders one result container', () => {
        const { container } = render(<FileMatch {...defaultProps} />)
        expect(getByTestId(container, 'result-container')).toBeTruthy()
        expect(getAllByTestId(container, 'result-container').length).toBe(1)
    })

    // Tests below should eventually go in a separate file, ResultsContainer.test.tsx
    it('correctly displays the repository name (without github.com) in the search result header', () => {
        const { container } = render(<FileMatch {...defaultProps} />)
        const header = getByTestId(container, 'result-container-header')
        expect(header).toBeTruthy()
        expect(getByText(header, 'golang/oauth2')).toBeTruthy()
    })

    it('displays the file path of the search result', () => {
        const { container } = render(<FileMatch {...defaultProps} />)
        const header = getByTestId(container, 'result-container-header')
        expect(header).toBeTruthy()
        expect(getByText(header, '.travis.yml')).toBeTruthy()
    })
})
