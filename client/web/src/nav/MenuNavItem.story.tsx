import { storiesOf } from '@storybook/react'
import React from 'react'

import { BatchChangesNavItem } from '../batches/BatchChangesNavItem'
import { CodeMonitoringNavItem } from '../code-monitoring/CodeMonitoringNavItem'
import { WebStory } from '../components/WebStory'
import { InsightsNavItem } from '../insights/components/InsightsNavLink/InsightsNavLink'

import { MenuNavItem } from './MenuNavItem'

const { add } = storiesOf('web/nav/MenuNavItem', module)

add(
    'Menu',
    () => (
        <WebStory>
            {() => (
                <MenuNavItem openByDefault={true}>
                    <BatchChangesNavItem isSourcegraphDotCom={false} />
                    <InsightsNavItem />
                    <CodeMonitoringNavItem />
                </MenuNavItem>
            )}
        </WebStory>
    ),
    {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/SFhXbl23TJ2j5tOF51NDtF/%F0%9F%93%9AWeb?node-id=1108%3A872',
        },
    }
)
