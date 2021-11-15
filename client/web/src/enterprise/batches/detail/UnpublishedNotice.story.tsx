import { storiesOf } from '@storybook/react'
import React from 'react'

import { WebStory } from '../../../components/WebStory'

import { UnpublishedNotice } from './UnpublishedNotice'

const { add } = storiesOf('web/batches/details/UnpublishedNotice', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('None published', () => <WebStory>{() => <UnpublishedNotice unpublished={10} total={10} />}</WebStory>)
