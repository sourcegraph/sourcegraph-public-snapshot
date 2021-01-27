import { storiesOf } from '@storybook/react'
import React from 'react'
import { WebStory } from '../components/WebStory'
import { CampaignsNavItem } from '../enterprise/campaigns/global/nav/CampaignsNavItem'
import { CodeMonitoringNavItem } from '../enterprise/code-monitoring/CodeMonitoringNavItem'
import { InsightsNavItem } from '../insights/InsightsNavLink'
import { MenuNavItem } from './MenuNavItem'

const { add } = storiesOf('web/nav/MenuNavItem', module)

add(
    'Menu',
    () => (
        <WebStory>
            {() => (
                <MenuNavItem openByDefault={true}>
                    <CampaignsNavItem />
                    <InsightsNavItem />
                    <CodeMonitoringNavItem />
                </MenuNavItem>
            )}
        </WebStory>
    ),
    {
        design: {
            type: 'figma',
            url:
                'https://www.figma.com/file/QSEhj2Nt4MSYLTjqVa2mOO/%2316567-Navigation-short-term-updates?node-id=12%3A299',
        },
    }
)
