import { storiesOf } from '@storybook/react'
import React from 'react'
import { CampaignsNavItem } from '../enterprise/campaigns/global/nav/CampaignsNavItem'
import { CodeMonitoringNavItem } from '../enterprise/code-monitoring/CodeMonitoringNavItem'
import { InsightsNavItem } from '../insights/InsightsNavLink'
import { MenuNavItem } from './MenuNavItem'

const { add } = storiesOf('web/nav/MenuNavItem', module)

add(
    'Menu',
    () => (
        <MenuNavItem>
            <CampaignsNavItem />
            <InsightsNavItem />
            <CodeMonitoringNavItem />
        </MenuNavItem>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/QSEhj2Nt4MSYLTjqVa2mOO/%2316567-Navigation-short-term-updates?node-id=12%3A299',
        },
    }
)
