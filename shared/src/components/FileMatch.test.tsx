import { createBrowserHistory } from 'history'
import FileIcon from 'mdi-react/FileIcon'
import * as React from 'react'
import { cleanup, getAllByTestId, getByTestId, getByText, render } from 'react-testing-library'
import _VisibilitySensor from 'react-visibility-sensor'
import sinon from 'sinon'
import { HIGHLIGHTED_FILE_LINES_REQUEST, NOOP_SETTINGS_CASCADE, RESULT } from '../../../web/src/search/testHelpers'
import { MockVisibilitySensor } from './CodeExcerpt.test'
import { FileMatch, IFileMatch } from './FileMatch'
import { setLinkComponent } from './Link'

jest.mock(
    'react-visibility-sensor',
    (): typeof _VisibilitySensor => ({ children, onChange }) => (
        <>
            <MockVisibilitySensor onChange={onChange} children={children} />
        </>
    )
)

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
        fetchHighlightedFileLines: HIGHLIGHTED_FILE_LINES_REQUEST,
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
