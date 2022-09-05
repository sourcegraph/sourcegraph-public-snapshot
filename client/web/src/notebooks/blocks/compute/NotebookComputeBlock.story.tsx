import { Meta, Story, DecoratorFn } from '@storybook/react'
import { noop } from 'lodash'

import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { WebStory } from '../../../components/WebStory'

import { NotebookComputeBlock } from './NotebookComputeBlock'

const decorator: DecoratorFn = story => <div className="p-3 container">{story()}</div>

const config: Meta = {
    title: 'web/search/notebooks/blocks/compute/NotebookComputeBlock',
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
}

export const Default: Story = () => (
    <WebStory>
        {props => (
            <NotebookComputeBlock
                {...props}
                {...noopBlockCallbacks}
                input=""
                output=""
                id="compute-block-1"
                isSelected={true}
                isReadOnly={false}
                isOtherBlockSelected={false}
                platformContext={NOOP_PLATFORM_CONTEXT}
            />
        )}
    </WebStory>
)
