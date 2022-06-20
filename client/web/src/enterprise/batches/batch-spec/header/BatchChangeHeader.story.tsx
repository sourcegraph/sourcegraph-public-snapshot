import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../../components/WebStory'

import { BatchChangeHeader } from './BatchChangeHeader'

const { add } = storiesOf('web/batches/batch-spec/header/BatchChangeHeader', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('creating a new batch change', () => (
    <WebStory>{props => <BatchChangeHeader {...props} title={{ text: 'Create batch change' }} />}</WebStory>
))

add('batch change already exists', () => (
    <WebStory>
        {props => (
            <BatchChangeHeader
                {...props}
                namespace={{ to: '/users/my-username', text: 'my-username' }}
                title={{ to: '/users/my-username/batch-changes/my-batch-change', text: 'my-batch-change' }}
                description="This is a description of my batch change."
            />
        )}
    </WebStory>
))
