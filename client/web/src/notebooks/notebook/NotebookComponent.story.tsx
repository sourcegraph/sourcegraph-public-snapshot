import { storiesOf } from '@storybook/react'
import { NEVER, of } from 'rxjs'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    extensionsController,
    HIGHLIGHTED_FILE_LINES_LONG,
    NOOP_PLATFORM_CONTEXT,
} from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { BlockInit } from '..'
import { WebStory } from '../../components/WebStory'

import { NotebookComponent } from './NotebookComponent'

const { add } = storiesOf('web/search/notebooks/notebook/NotebookComponent', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({ chromatic: { disableSnapshots: false } })

const blocks: BlockInit[] = [
    { id: '1', type: 'md', input: { text: '# Markdown' } },
    { id: '2', type: 'query', input: { query: 'Query' } },
    { id: '3', type: 'md', input: { text: '# Markdown 1' } },
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

add('default', () => (
    <WebStory>
        {props => (
            <NotebookComponent
                {...props}
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
                authenticatedUser={null}
                showSearchContext={true}
                platformContext={NOOP_PLATFORM_CONTEXT}
                exportedFileName="notebook.snb.md"
                onCopyNotebook={() => NEVER}
            />
        )}
    </WebStory>
))

add('default read-only', () => (
    <WebStory>
        {props => (
            <NotebookComponent
                {...props}
                isReadOnly={true}
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
                authenticatedUser={null}
                showSearchContext={true}
                platformContext={NOOP_PLATFORM_CONTEXT}
                exportedFileName="notebook.snb.md"
                onCopyNotebook={() => NEVER}
            />
        )}
    </WebStory>
))
