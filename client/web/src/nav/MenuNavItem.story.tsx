import { Meta, Story } from '@storybook/react'
import BarChartIcon from 'mdi-react/BarChartIcon'
import React from 'react'

import { BatchChangesNavItem } from '../batches/BatchChangesNavItem'
import { CodeMonitoringNavItem } from '../code-monitoring/CodeMonitoringNavItem'
import { LinkWithIcon } from '../components/LinkWithIcon'
import { WebStory } from '../components/WebStory'

import { MenuNavItem } from './MenuNavItem'

const config: Meta = {
    title: 'web/nav/MenuNavItem',
    decorators: [story => <div className="p-3 container h-100 web-content">{story()}</div>],
    parameters: {
        design: {
            type: 'figma',
            url: 'https://www.figma.com/file/SFhXbl23TJ2j5tOF51NDtF/%F0%9F%93%9AWeb?node-id=1108%3A872',
        },
        chromatic: {
            enableDarkMode: true,
            viewports: [400],
        },
    },
}

export default config

const InsightsNavItem: React.FunctionComponent = () => (
    <LinkWithIcon
        to="/insights"
        text="Insights"
        icon={BarChartIcon}
        className="nav-link text-decoration-none"
        activeClassName="active"
    />
)

export const Menu: Story = () => (
    <WebStory>
        {() => (
            <MenuNavItem openByDefault={true}>
                <BatchChangesNavItem />
                <InsightsNavItem />
                <CodeMonitoringNavItem />
            </MenuNavItem>
        )}
    </WebStory>
)
