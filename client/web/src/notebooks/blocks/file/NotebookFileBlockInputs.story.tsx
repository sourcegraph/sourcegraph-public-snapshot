import { Meta, Story, DecoratorFn } from '@storybook/react'
import { noop } from 'lodash'

import { WebStory } from '../../../components/WebStory'

import { NotebookFileBlockInputs } from './NotebookFileBlockInputs'

const decorator: DecoratorFn = story => <div className="container p-3">{story()}</div>

const config: Meta = {
    title: 'web/search/notebooks/blocks/file/NotebookFileBlockInputs',
    decorators: [decorator],
}

export default config

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

export const Default: Story = () => (
    <WebStory>{webProps => <NotebookFileBlockInputs {...webProps} {...defaultProps} />}</WebStory>
)
