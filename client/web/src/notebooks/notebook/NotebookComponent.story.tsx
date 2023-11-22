import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { NEVER, of } from 'rxjs'

import { EMPTY_SETTINGS_CASCADE } from '@sourcegraph/shared/src/settings/settings'
import { NOOP_TELEMETRY_SERVICE } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { HIGHLIGHTED_FILE_LINES_LONG, NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import type { BlockInit } from '..'
import { WebStory } from '../../components/WebStory'

import { NotebookComponent } from './NotebookComponent'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/search/notebooks/notebook/NotebookComponent',
    parameters: {
        chromatic: { disableSnapshots: false },
    },
    decorators: [decorator],
}

export default config

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

export const Default: StoryFn = () => (
    <WebStory>
        {props => (
            <NotebookComponent
                {...props}
                isSourcegraphDotCom={true}
                searchContextsEnabled={true}
                ownEnabled={true}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                streamSearch={() => NEVER}
                fetchHighlightedFileLineRanges={() => of(HIGHLIGHTED_FILE_LINES_LONG)}
                onSerializeBlocks={() => {}}
                blocks={blocks}
                settingsCascade={EMPTY_SETTINGS_CASCADE}
                authenticatedUser={null}
                platformContext={NOOP_PLATFORM_CONTEXT}
                exportedFileName="notebook.snb.md"
                onCopyNotebook={() => NEVER}
            />
        )}
    </WebStory>
)

export const DefaultReadOnly: StoryFn = () => (
    <WebStory>
        {props => (
            <NotebookComponent
                {...props}
                isReadOnly={true}
                isSourcegraphDotCom={true}
                searchContextsEnabled={true}
                ownEnabled={true}
                telemetryService={NOOP_TELEMETRY_SERVICE}
                streamSearch={() => NEVER}
                fetchHighlightedFileLineRanges={() => of(HIGHLIGHTED_FILE_LINES_LONG)}
                onSerializeBlocks={() => {}}
                blocks={blocks}
                settingsCascade={EMPTY_SETTINGS_CASCADE}
                authenticatedUser={null}
                platformContext={NOOP_PLATFORM_CONTEXT}
                exportedFileName="notebook.snb.md"
                onCopyNotebook={() => NEVER}
            />
        )}
    </WebStory>
)

DefaultReadOnly.storyName = 'default read-only'
