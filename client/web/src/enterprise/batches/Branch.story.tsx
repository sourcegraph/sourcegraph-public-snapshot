import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../components/WebStory'

import { BranchMerge } from './Branch'

const { add } = storiesOf('web/batches/Branch', module)

add('Forked', () => <WebStory>{() => <BranchMerge baseRef="main" forkNamespace="fork" headRef="branch" />}</WebStory>)

add('Will be forked into the user', () => (
    <WebStory>{() => <BranchMerge baseRef="main" forkNamespace="<user>" headRef="branch" />}</WebStory>
))

add('Unforked', () => <WebStory>{() => <BranchMerge baseRef="main" headRef="branch" />}</WebStory>)
