import { storiesOf } from '@storybook/react'
import { capitalize } from 'lodash'
import React from 'react'

import { ChangesetReviewState } from '../../../../graphql-operations'
import { EnterpriseWebStory } from '../../../components/EnterpriseWebStory'

import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'

const { add } = storiesOf('web/batches/ChangesetReviewStatusCell', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

for (const state of Object.values(ChangesetReviewState)) {
    add(capitalize(state), () => (
        <EnterpriseWebStory>{props => <ChangesetReviewStatusCell {...props} reviewState={state} />}</EnterpriseWebStory>
    ))
}
