import { storiesOf } from '@storybook/react'
import { createBrowserHistory } from 'history'
import FileIcon from 'mdi-react/FileIcon'
import * as React from 'react'
import _VisibilitySensor from 'react-visibility-sensor'
import sinon from 'sinon'

import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { SymbolKind } from '../graphql/schema'
import { HIGHLIGHTED_FILE_LINES_REQUEST, NOOP_SETTINGS_CASCADE } from '../util/searchTestHelpers'

import { FileMatch } from './FileMatch'

const defaultProps: Omit<React.ComponentProps<typeof FileMatch>, 'result'> = {
    location: createBrowserHistory().location,
    icon: FileIcon,
    onSelect: sinon.spy(),
    expanded: true,
    showAllMatches: true,
    isLightTheme: true,
    fetchHighlightedFileLineRanges: HIGHLIGHTED_FILE_LINES_REQUEST,
    settingsCascade: NOOP_SETTINGS_CASCADE,
}

// The FileMatch component lives in `shared`, but show it in the `web` storybook group so it's
// alongside the other search result component storybooks.
const { add } = storiesOf('web/search/results/FileMatch', module)
    .addParameters({
        chromatic: { viewports: [800] },
    })
    .addDecorator(story => (
        <>
            <div className="p-4">{story()}</div>
            <style>{webStyles}</style>
        </>
    ))

add('file/line matches', () => (
    <FileMatch
        {...defaultProps}
        result={{
            type: 'file',
            repository: 'a/b',
            name: '.foo.yml',
            lineMatches: [{ line: '', lineNumber: 1, offsetAndLengths: [[1, 1]] }],
        }}
    />
))

add('file/symbol matches', () => (
    <FileMatch
        {...defaultProps}
        result={{
            type: 'symbol',
            repository: 'a/b',
            name: '.foo.yml',
            symbols: [{ containerName: 'Cat', name: 'meow', kind: SymbolKind.METHOD, url: '/meow' }],
        }}
    />
))

add('file/path match', () => (
    <FileMatch
        {...defaultProps}
        result={{
            type: 'file',
            repository: 'a/b',
            name: '.foo.yml',
            lineMatches: [],
        }}
    />
))
