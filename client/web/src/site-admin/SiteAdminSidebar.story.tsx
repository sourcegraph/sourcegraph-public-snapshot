import { DecoratorFn, Meta, Story } from '@storybook/react'

import { WebStory } from '../components/WebStory'

import { siteAdminSidebarGroups } from './sidebaritems'
import { SiteAdminSidebar } from './SiteAdminSidebar'

const decorator: DecoratorFn = story => <div style={{ width: '192px' }}>{story()}</div>

const config: Meta = {
    title: 'web/site-admin/AdminSidebar',
    decorators: [decorator],
}

export default config

export const AdminSidebarItems: Story = () => (
    <WebStory>
        {webProps => (
            <SiteAdminSidebar
                {...webProps}
                groups={siteAdminSidebarGroups}
                isSourcegraphDotCom={false}
                batchChangesEnabled={false}
                batchChangesExecutionEnabled={false}
                batchChangesWebhookLogsEnabled={false}
            />
        )}
    </WebStory>
)

AdminSidebarItems.storyName = 'Admin Sidebar Items'
AdminSidebarItems.parameters = {
    design: {
        type: 'figma',
        url:
            'https://www.figma.com/file/cT23UcGqbdjdV4H2yxweIu/%2311775-Map-the-current-information-architecture-%5BApproved%5D?node-id=68%3A1',
    },
}
