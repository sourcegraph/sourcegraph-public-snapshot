import { storiesOf } from '@storybook/react'
import React from 'react'
import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'
import { ClosedNotice } from './ClosedNotice'

const { add } = storiesOf('web/campaigns/details/ClosedNotice', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

add('Campaign closed', () => <EnterpriseWebStory>{() => <ClosedNotice closedAt="2021-02-02" />}</EnterpriseWebStory>)
