import { storiesOf } from '@storybook/react'
import { radios, boolean } from '@storybook/addon-knobs'
import React from 'react'
import { CampaignNode } from './CampaignNode'
import { createMemoryHistory } from 'history'
import webStyles from '../../../enterprise.scss'
import { Tooltip } from '../../../components/tooltip/Tooltip'
import isChromatic from 'chromatic/isChromatic'
import { ListCampaign } from '../../../graphql-operations'
import { subDays } from 'date-fns'

export const nodes: Record<string, ListCampaign> = {
    'Open campaign': {
        id: 'test',
        name: 'Awesome campaign',
        description: `# What this does

This is my thorough explanation. And it can also get very long, in that case the UI doesn't break though, which is good. And one more line to finally be longer than the viewport.`,
        createdAt: subDays(new Date(), 5).toISOString(),
        closedAt: null,
        changesets: {
            stats: {
                open: 10,
                closed: 0,
                merged: 5,
            },
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    },
    'No description': {
        id: 'test2',
        name: 'Awesome campaign',
        description: null,
        createdAt: subDays(new Date(), 5).toISOString(),
        closedAt: null,
        changesets: {
            stats: {
                open: 10,
                closed: 0,
                merged: 5,
            },
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    },
    'Closed campaign': {
        id: 'test3',
        name: 'Awesome campaign',
        description: `# My campaign

        This is my thorough explanation.`,
        createdAt: subDays(new Date(), 5).toISOString(),
        closedAt: subDays(new Date(), 3).toISOString(),
        changesets: {
            stats: {
                open: 0,
                closed: 10,
                merged: 5,
            },
        },
        namespace: {
            namespaceName: 'alice',
            url: '/users/alice',
        },
    },
}

const { add } = storiesOf('web/campaigns/CampaignNode', module).addDecorator(story => {
    const theme = radios('Theme', { Light: 'light', Dark: 'dark' }, 'light')
    document.body.classList.toggle('theme-light', theme === 'light')
    document.body.classList.toggle('theme-dark', theme === 'dark')
    return (
        <>
            <Tooltip />
            <style>{webStyles}</style>
            <div className="p-3 container web-content campaign-list-page__grid">{story()}</div>
        </>
    )
})

for (const key of Object.keys(nodes)) {
    add(key, () => (
        <CampaignNode
            node={nodes[key]}
            displayNamespace={boolean('Display namespace', true)}
            now={isChromatic() ? subDays(new Date(), 5) : undefined}
            history={createMemoryHistory()}
        />
    ))
}
