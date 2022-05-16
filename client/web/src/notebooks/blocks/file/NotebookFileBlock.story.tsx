import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import { of } from 'rxjs'

import { extensionsController, HIGHLIGHTED_FILE_LINES_LONG } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { FileBlockInput } from '../..'
import { WebStory } from '../../../components/WebStory'

import { NotebookFileBlock } from './NotebookFileBlock'

const { add } = storiesOf('web/search/notebooks/blocks/file/NotebookFileBlock', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

const noopBlockCallbacks = {
    onRunBlock: noop,
    onBlockInputChange: noop,
    onSelectBlock: noop,
    onMoveBlockSelection: noop,
    onDeleteBlock: noop,
    onMoveBlock: noop,
    onDuplicateBlock: noop,
}

const fileBlockInput: FileBlockInput = {
    repositoryName: 'github.com/sourcegraph/sourcegraph',
    filePath: 'client/web/file.tsx',
    revision: 'main',
    lineRange: null,
}

add('default', () => (
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
                isOtherBlockSelected={false}
                isSourcegraphDotCom={false}
                extensionsController={extensionsController}
                sourcegraphSearchLanguageId="sourcegraph"
            />
        )}
    </WebStory>
))

add('edit mode', () => (
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
                isOtherBlockSelected={false}
                isSourcegraphDotCom={false}
                extensionsController={extensionsController}
                sourcegraphSearchLanguageId="sourcegraph"
            />
        )}
    </WebStory>
))

add('error fetching file', () => (
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
                isOtherBlockSelected={false}
                isSourcegraphDotCom={false}
                extensionsController={extensionsController}
                sourcegraphSearchLanguageId="sourcegraph"
            />
        )}
    </WebStory>
))
