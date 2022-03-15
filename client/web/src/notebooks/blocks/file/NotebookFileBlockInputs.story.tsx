import React from 'react'

import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'

import { NotebookFileBlockInputs } from './NotebookFileBlockInputs'

const { add } = storiesOf('web/search/notebooks/blocks/file/NotebookFileBlockInputs', module).addDecorator(story => (
    <div className="container p-3">{story()}</div>
))

const defaultProps = {
    id: 'block-id',
    showRevisionInput: true,
    showLineRangeInput: true,
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

add('default', () => <WebStory>{() => <NotebookFileBlockInputs {...defaultProps} />}</WebStory>)

add('all valid', () => (
    <WebStory>
        {() => (
            <NotebookFileBlockInputs
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
            <NotebookFileBlockInputs
                {...defaultProps}
                isRepositoryNameValid={false}
                isFilePathValid={false}
                isRevisionValid={false}
                isLineRangeValid={false}
            />
        )}
    </WebStory>
))
