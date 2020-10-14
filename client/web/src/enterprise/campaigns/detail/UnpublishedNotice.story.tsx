import { storiesOf } from '@storybook/react'
import React from 'react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { UnpublishedNotice } from './UnpublishedNotice'

const { add } = storiesOf('web/campaigns/details/UnpublishedNotice', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Some unpublished', () => <EnterpriseWebStory>{() => <UnpublishedNotice unpublished={10} />}</EnterpriseWebStory>)
