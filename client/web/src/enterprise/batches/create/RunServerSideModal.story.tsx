import React from 'react'

import { storiesOf } from '@storybook/react'

import { WebStory } from '../../../components/WebStory'

import { RunServerSideModal } from './RunServerSideModal'

const { add } = storiesOf('web/batches/create/RunServerSideModal', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Download Spec Modal', () => (
    <WebStory>{props => <RunServerSideModal name="" originalInput="" {...props} />}</WebStory>
))
