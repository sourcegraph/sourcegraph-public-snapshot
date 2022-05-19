import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../../components/WebStory'

import { LibraryPane } from './LibraryPane'

const { add } = storiesOf('web/batches/batch-spec/edit/LibraryPane', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('editable', () => (
    <WebStory>
        {props => <LibraryPane {...props} name="my-batch-change" onReplaceItem={() => alert('batch spec replaced!')} />}
    </WebStory>
))

add('read-only', () => (
    <WebStory>{props => <LibraryPane {...props} name="my-batch-change" isReadOnly={true} />}</WebStory>
))
