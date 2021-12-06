import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { SearchNotebookFileBlockInputs } from './SearchNotebookFileBlockInputs'

const { add } = storiesOf('web/search/notebook/fileBlock/SearchNotebookFileBlockInputs', module).addDecorator(story => (
    <div className="container p-3">{story()}</div>
))

const defaultProps = {
    id: 'block-id',
    showRevisionInput: true,
    showLineRangeInput: true,
    isMacPlatform: false,
    setIsInputFocused: noop,
    setFileInput: noop,
    setLineRangeInput: noop,
    onSelectBlock: noop,
    isRepositoryNameValid: undefined,
    isFilePathValid: undefined,
    isRevisionValid: undefined,
    isLineRangeValid: undefined,
    repositoryName: 'github.com/sourcegraph/sourcegraph',
    revision: 'main',
    filePath: 'client/web/file.tsx',
    lineRangeInput: '123-321',
}

add('default', () => <WebStory>{() => <SearchNotebookFileBlockInputs {...defaultProps} />}</WebStory>)

add('all valid', () => (
    <WebStory>
        {() => (
            <SearchNotebookFileBlockInputs
                {...defaultProps}
                isRepositoryNameValid={true}
                isFilePathValid={true}
                isRevisionValid={true}
                isLineRangeValid={true}
            />
        )}
    </WebStory>
))

add('all invalid', () => (
    <WebStory>
        {() => (
            <SearchNotebookFileBlockInputs
                {...defaultProps}
                isRepositoryNameValid={false}
                isFilePathValid={false}
                isRevisionValid={false}
                isLineRangeValid={false}
            />
        )}
    </WebStory>
))
