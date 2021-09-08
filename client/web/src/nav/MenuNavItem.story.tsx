import { storiesOf } from '@storybook/react'
import BarChartIcon from 'mdi-react/BarChartIcon'
import React from 'react'

import { BatchChangesNavItem } from '../batches/BatchChangesNavItem'
import { CodeMonitoringNavItem } from '../code-monitoring/CodeMonitoringNavItem'
import { LinkWithIcon } from '../components/LinkWithIcon'
import { WebStory } from '../components/WebStory'

import { MenuNavItem } from './MenuNavItem'

const InsightsNavItem: React.FunctionComponent = () => (
    <LinkWithIcon
        to="/insights"
        text="Insights"
        icon={BarChartIcon}
        className="nav-link btn btn-link text-decoration-none"
        activeClassName="active"
    />
)

const { add } = storiesOf('web/nav/MenuNavItem', module)

add(
    'Menu',
    () => (
        <WebStory>
            {() => (
                <MenuNavItem openByDefault={true}>
                    <BatchChangesNavItem />
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
