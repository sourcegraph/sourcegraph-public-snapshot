import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { LibraryPane } from './LibraryPane'

const { add } = storiesOf('web/batches/batch-spec/edit', module)
    .addDecorator(story => <div className="p-3 container">{story()}</div>)
    .addParameters({
        chromatic: {
            disableSnapshot: false,
        },
    })

add('LibraryPane', () => (
    <WebStory>
        {props => <LibraryPane {...props} name="my-batch-change" onReplaceItem={() => alert('batch spec replaced!')} />}
    </WebStory>
))
