import { storiesOf } from '@storybook/react'
import { noop } from 'lodash'

import { NOOP_PLATFORM_CONTEXT } from '@sourcegraph/shared/src/testing/searchTestHelpers'

import { WebStory } from '../../../components/WebStory'

import { NotebookComputeBlock } from './NotebookComputeBlock'

const { add } = storiesOf('web/search/notebooks/blocks/compute/NotebookComputeBlock', module).addDecorator(story => (
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

add('default', () => (
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
))
