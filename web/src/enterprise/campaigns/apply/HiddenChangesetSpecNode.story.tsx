import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { HiddenChangesetSpecNode } from './HiddenChangesetSpecNode'
import { addDays } from 'date-fns'
import { ChangesetSpecType } from '../../../../../shared/src/graphql/schema'
import { ChangesetSpecFields } from '../../../graphql-operations'

const { add } = storiesOf('web/campaigns/apply/HiddenChangesetSpecNode', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container web-content changeset-spec-list__grid">{story()}</div>
        </>
    )
})

export const hiddenChangesetSpecStories: Record<string, ChangesetSpecFields & { __typename: 'HiddenChangesetSpec' }> = {
    'Import changeset': {
        __typename: 'HiddenChangesetSpec',
        id: 'someidh1',
        expiresAt: addDays(new Date(), 7).toISOString(),
        type: ChangesetSpecType.EXISTING,
    },
    'Create changeset': {
        __typename: 'HiddenChangesetSpec',
        id: 'someidh2',
        expiresAt: addDays(new Date(), 7).toISOString(),
        type: ChangesetSpecType.BRANCH,
    },
}

for (const storyName of Object.keys(hiddenChangesetSpecStories)) {
    add(storyName, () => <HiddenChangesetSpecNode node={hiddenChangesetSpecStories[storyName]} />)
}
