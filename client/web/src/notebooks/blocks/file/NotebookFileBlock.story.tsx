import { DecoratorFn, Meta, Story } from '@storybook/react'
import { noop } from 'lodash'
import { of } from 'rxjs'

import { HIGHLIGHTED_FILE_LINES_LONG } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { FileBlockInput } from '../..'
import { WebStory } from '../../../components/WebStory'

import { NotebookFileBlock } from './NotebookFileBlock'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

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

export const Default: Story = () => (
    <WebStory>
        {props => (
            <NotebookFileBlock
                {...props}
                {...noopBlockCallbacks}
                id="file-block-1"
                input={fileBlockInput}
                output={of(HIGHLIGHTED_FILE_LINES_LONG[0])}
                isSelected={true}
                isReadOnly={false}
                showMenu={false}
                isSourcegraphDotCom={false}
            />
        )}
    </WebStory>
)

export const EditMode: Story = () => (
    <WebStory>
        {props => (
            <NotebookFileBlock
                {...props}
                {...noopBlockCallbacks}
                id="file-block-1"
                input={{ repositoryName: '', filePath: '', revision: 'main', lineRange: { startLine: 1, endLine: 10 } }}
                output={of(HIGHLIGHTED_FILE_LINES_LONG[0])}
                isSelected={true}
                isReadOnly={false}
                showMenu={false}
                isSourcegraphDotCom={false}
            />
        )}
    </WebStory>
)

EditMode.storyName = 'edit mode'

export const ErrorFetchingFile: Story = () => (
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
            />
        )}
    </WebStory>
)

ErrorFetchingFile.storyName = 'error fetching file'
