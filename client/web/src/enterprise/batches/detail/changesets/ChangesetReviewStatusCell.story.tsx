import { storiesOf } from '@storybook/react'
import { capitalize } from 'lodash'
import React from 'react'

import { WebStory } from '../../../../components/WebStory'
import { ChangesetReviewState } from '../../../../graphql-operations'

import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'

const { add } = storiesOf('web/batches/ChangesetReviewStatusCell', module).addDecorator(story => (
    <div className="p-3 container">{story()}</div>
))

for (const state of Object.values(ChangesetReviewState)) {
    add(capitalize(state), () => (
        <WebStory>{props => <ChangesetReviewStatusCell {...props} reviewState={state} />}</WebStory>
    ))
}
