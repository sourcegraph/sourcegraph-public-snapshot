import { storiesOf } from '@storybook/react'
import React from 'react'
import webStyles from '../../../../enterprise.scss'
import { ChangesetReviewStatusCell } from './ChangesetReviewStatusCell'
import { ChangesetReviewState } from '../../../../graphql-operations'
import { capitalize } from 'lodash'
import { WebStory } from '../../../../components/WebStory'

const { add } = storiesOf('web/campaigns/ChangesetReviewStatusCell', module).addDecorator(story => (
    <div className="p-3 container web-content">{story()}</div>
))

for (const state of Object.values(ChangesetReviewState)) {
    add(capitalize(state), () => (
        <WebStory webStyles={webStyles}>
            {props => <ChangesetReviewStatusCell {...props} reviewState={state} />}
        </WebStory>
    ))
}
