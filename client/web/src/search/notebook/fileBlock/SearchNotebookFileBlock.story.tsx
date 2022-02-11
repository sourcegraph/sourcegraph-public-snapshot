import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'
import { of } from 'rxjs'

import { HIGHLIGHTED_FILE_LINES_LONG } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { FileBlockInput } from '..'
import { WebStory } from '../../../components/WebStory'
import { RepositoryFields } from '../../../graphql-operations'

import { SearchNotebookFileBlock } from './SearchNotebookFileBlock'

const { add } = storiesOf('web/search/notebook/fileBlock/SearchNotebookFileBlock', module).addDecorator(story => (
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

const resolveRevision = () => of({ commitID: 'commit1', defaultBranch: 'main', rootTreeURL: '' })
const fetchRepository = () => of({ id: 'repo' } as RepositoryFields)

add('default', () => (
    <WebStory>
        {props => (
            <SearchNotebookFileBlock
                {...props}
                {...noopBlockCallbacks}
                id="file-block-1"
                input={fileBlockInput}
                output={of(HIGHLIGHTED_FILE_LINES_LONG[0])}
                isSelected={true}
                isReadOnly={false}
                isOtherBlockSelected={false}
                isMacPlatform={true}
                isSourcegraphDotCom={false}
                fetchHighlightedFileLineRanges={() => of(HIGHLIGHTED_FILE_LINES_LONG)}
                resolveRevision={resolveRevision}
                fetchRepository={fetchRepository}
            />
        )}
    </WebStory>
))

add('edit mode', () => (
    <WebStory>
        {props => (
            <SearchNotebookFileBlock
                {...props}
                {...noopBlockCallbacks}
                id="file-block-1"
                input={{ repositoryName: '', filePath: '', revision: 'main', lineRange: { startLine: 1, endLine: 10 } }}
                output={of(HIGHLIGHTED_FILE_LINES_LONG[0])}
                isSelected={true}
                isReadOnly={false}
                isOtherBlockSelected={false}
                isMacPlatform={true}
                isSourcegraphDotCom={false}
                fetchHighlightedFileLineRanges={() => of(HIGHLIGHTED_FILE_LINES_LONG)}
                resolveRevision={resolveRevision}
                fetchRepository={fetchRepository}
            />
        )}
    </WebStory>
))

add('error fetching file', () => (
    <WebStory>
        {props => (
            <SearchNotebookFileBlock
                {...props}
                {...noopBlockCallbacks}
                id="file-block-1"
                input={fileBlockInput}
                output={of(new Error('File not found'))}
                isSelected={true}
                isReadOnly={false}
                isOtherBlockSelected={false}
                isMacPlatform={true}
                isSourcegraphDotCom={false}
                fetchHighlightedFileLineRanges={() => of(HIGHLIGHTED_FILE_LINES_LONG)}
                resolveRevision={resolveRevision}
                fetchRepository={fetchRepository}
            />
        )}
    </WebStory>
))
