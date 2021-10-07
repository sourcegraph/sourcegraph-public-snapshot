import { storiesOf } from '@storybook/react'
import React from 'react'

import { EnterpriseWebStory } from '../../components/EnterpriseWebStory'

import { DotcomGettingStartedPage } from './DotcomGettingStartedPage'

const { add } = storiesOf('web/batches/DotcomGettingStartedPage', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

add('Overview', () => <EnterpriseWebStory>{() => <DotcomGettingStartedPage />}</EnterpriseWebStory>)
