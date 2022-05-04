import { storiesOf } from '@storybook/react'

import { WebStory } from '../../components/WebStory'

import { BranchMerge } from './Branch'

const { add } = storiesOf('web/batches/Branch', module)

add('Forked', () => (
    <WebStory>
        {() => <BranchMerge baseRef="main" forkTarget={{ pushUser: false, namespace: 'org' }} headRef="branch" />}
    </WebStory>
))

add('Will be forked into the user', () => (
    <WebStory>
        {() => <BranchMerge baseRef="main" forkTarget={{ pushUser: true, namespace: 'org' }} headRef="branch" />}
    </WebStory>
))

add('Unforked', () => <WebStory>{() => <BranchMerge baseRef="main" headRef="branch" />}</WebStory>)
