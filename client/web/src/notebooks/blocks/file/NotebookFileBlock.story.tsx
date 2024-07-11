import type { Decorator, Meta, StoryFn } from '@storybook/react'
import { noop } from 'lodash'
import { of } from 'rxjs'

import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { HIGHLIGHTED_FILE_LINES_LONG } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import type { FileBlockInput } from '../..'
import { WebStory } from '../../../components/WebStory'

import { NotebookFileBlock } from './NotebookFileBlock'

const decorator: Decorator = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/search/notebooks/blocks/file/NotebookFileBlock',
    decorators: [decorator],
}

export default config

const noopBlockCallbacks = {
    onRunBlock: noop,
    onBlockInputChange: noop,
    onSelectBlock: noop,
    onMoveBlockSelection: noop,
    onDeleteBlock: noop,
    onMoveBlock: noop,
    onDuplicateBlock: noop,
    onNewBlock: noop,
}

const fileBlockInput: FileBlockInput = {
    repositoryName: 'github.com/sourcegraph/sourcegraph',
    filePath: 'client/web/file.tsx',
    revision: 'main',
    lineRange: null,
}

export const Default: StoryFn = () => (
    <WebStory>
        {props => (
            <NotebookFileBlock
                {...props}
                {...noopBlockCallbacks}
                id="file-block-1"
                input={fileBlockInput}
                output={of(HIGHLIGHTED_FILE_LINES_LONG)}
                isSelected={true}
                isReadOnly={false}
                showMenu={false}
                isSourcegraphDotCom={false}
                patternType={SearchPatternType.standard}
            />
        )}
    </WebStory>
)

export const EditMode: StoryFn = () => (
    <WebStory>
        {props => (
            <NotebookFileBlock
                {...props}
                {...noopBlockCallbacks}
                id="file-block-1"
                input={{ repositoryName: '', filePath: '', revision: 'main', lineRange: { startLine: 1, endLine: 10 } }}
                output={of(HIGHLIGHTED_FILE_LINES_LONG)}
                isSelected={true}
                isReadOnly={false}
                showMenu={false}
                isSourcegraphDotCom={false}
                patternType={SearchPatternType.standard}
            />
        )}
    </WebStory>
)

EditMode.storyName = 'edit mode'

export const ErrorFetchingFile: StoryFn = () => (
    <WebStory>
        {props => (
            <NotebookFileBlock
                {...props}
                {...noopBlockCallbacks}
                id="file-block-1"
                input={fileBlockInput}
                output={of(new Error('File not found'))}
                isSelected={true}
                isReadOnly={false}
                showMenu={false}
                isSourcegraphDotCom={false}
                patternType={SearchPatternType.standard}
            />
        )}
    </WebStory>
)

ErrorFetchingFile.storyName = 'error fetching file'
