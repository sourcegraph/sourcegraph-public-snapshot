import { storiesOf } from '@storybook/react'
import React from 'react'
import { NEVER, of } from 'rxjs'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/client-api'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    extensionsController,
    HIGHLIGHTED_FILE_LINES_LONG,
    NOOP_PLATFORM_CONTEXT,
} from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { WebStory } from '../../components/WebStory'
import { RepositoryFields } from '../../graphql-operations'

import { SearchNotebook } from './SearchNotebook'

import { BlockInit } from '.'

const { add } = storiesOf('web/search/notebook/SearchNotebook', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({ chromatic: { disableSnapshots: false } })

const blocks: BlockInit[] = [
    { id: '1', type: 'md', input: '# Markdown' },
    { id: '2', type: 'query', input: 'Query' },
    { id: '3', type: 'md', input: '# Markdown 1' },
    {
        id: '4',
        type: 'file',
        input: {
            repositoryName: 'github.com/sourcegraph/sourcegraph',
            filePath: 'client/web/file.tsx',
            revision: 'main',
            lineRange: null,
        },
    },
]

const resolveRevision = () => of({ commitID: 'commit1', defaultBranch: 'main', rootTreeURL: '' })
const fetchRepository = () => of({ id: 'repo' } as RepositoryFields)

add('default', () => (
    <WebStory>
        {props => (
            <SearchNotebook
                {...props}
                isMacPlatform={true}
                isSourcegraphDotCom={true}
                searchContextsEnabled={true}
                globbing={true}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                streamSearch={() => NEVER}
                fetchHighlightedFileLineRanges={() => of(HIGHLIGHTED_FILE_LINES_LONG)}
                onSerializeBlocks={() => {}}
                blocks={blocks}
                settingsCascade={EMPTY_SETTINGS_CASCADE}
                extensionsController={extensionsController}
                fetchRepository={fetchRepository}
                resolveRevision={resolveRevision}
                authenticatedUser={null}
                showSearchContext={true}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )}
    </WebStory>
))

add('default read-only', () => (
    <WebStory>
        {props => (
            <SearchNotebook
                {...props}
                isReadOnly={true}
                isMacPlatform={true}
                isSourcegraphDotCom={true}
                searchContextsEnabled={true}
                globbing={true}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                streamSearch={() => NEVER}
                fetchHighlightedFileLineRanges={() => of(HIGHLIGHTED_FILE_LINES_LONG)}
                onSerializeBlocks={() => {}}
                blocks={blocks}
                settingsCascade={EMPTY_SETTINGS_CASCADE}
                extensionsController={extensionsController}
                fetchRepository={fetchRepository}
                resolveRevision={resolveRevision}
                authenticatedUser={null}
                showSearchContext={true}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )}
    </WebStory>
))
