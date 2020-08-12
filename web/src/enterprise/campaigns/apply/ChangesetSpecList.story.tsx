import * as H from 'history'
import { storiesOf } from '@storybook/react'
import { radios } from '@storybook/addon-knobs'
import React from 'react'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import { ChangesetSpecList } from './ChangesetSpecList'
import { of, Observable } from 'rxjs'
import { CampaignSpecChangesetSpecsResult, ChangesetSpecFields } from '../../../graphql-operations'
import { visibleChangesetSpecStories } from './VisibleChangesetSpecNode.story'
import { hiddenChangesetSpecStories } from './HiddenChangesetSpecNode.story'

let isLightTheme = true
const { add } = storiesOf('web/campaigns/apply/ChangesetSpecList', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    isLightTheme = theme === 'light'
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container web-content">{story()}</div>
        </>
    )
})

const nodes: ChangesetSpecFields[] = [
    ...Object.values(visibleChangesetSpecStories),
    ...Object.values(hiddenChangesetSpecStories),
]

const queryChangesetSpecs = (): Observable<
    (CampaignSpecChangesetSpecsResult['node'] & { __typename: 'CampaignSpec' })['changesetSpecs']
> =>
    of({
        pageInfo: {
            endCursor: null,
            hasNextPage: false,
        },
        totalCount: nodes.length,
        nodes,
    })

add('List view', () => {
    const history = H.createMemoryHistory()
    return (
        <ChangesetSpecList
            campaignSpecID="123123"
            queryChangesetSpecs={queryChangesetSpecs}
            history={history}
            location={history.location}
            isLightTheme={isLightTheme}
        />
    )
})
