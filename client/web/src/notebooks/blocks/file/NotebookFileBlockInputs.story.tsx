import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'

import { NotebookFileBlockInputs } from './NotebookFileBlockInputs'

const { add } = storiesOf('web/search/notebooks/blocks/file/NotebookFileBlockInputs', module).addDecorator(story => (
    <div className="container p-3">{story()}</div>
))

const defaultProps = {
    id: 'block-id',
    repositoryName: 'github.com/sourcegraph/sourcegraph',
    revision: 'main',
    filePath: 'client/web/file.tsx',
    lineRangeInput: '123-321',
    sourcegraphSearchLanguageId: 'sourcegraph',
    queryInput: '',
    setQueryInput: noop,
    debouncedSetQueryInput: noop,
    onFileSelected: noop,
    onRunBlock: noop,
    lineRange: null,
    onLineRangeChange: noop,
    editor: undefined,
    setEditor: noop,
}

add('default', () => <WebStory>{webProps => <NotebookFileBlockInputs {...webProps} {...defaultProps} />}</WebStory>)
