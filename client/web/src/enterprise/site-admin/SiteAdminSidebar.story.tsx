import type { DecoratorFn, Meta, Story } from '@storybook/react'

import { Code, Grid } from '@sourcegraph/wildcard'

import { WebStory } from '../../components/WebStory'
import { SiteAdminSidebar } from '../../site-admin/SiteAdminSidebar'

import { enterpriseSiteAdminSidebarGroups } from './sidebaritems'

const decorator: DecoratorFn = story => <div style={{ width: '192px' }}>{story()}</div>

const config: Meta = {
    title: 'web/site-admin/AdminSidebar',
    decorators: [decorator],
    parameters: {
        chromatic: {
            disableSnapshot: false,
        },
    },
}

export default config

// Moved story under enterprise folder to avoid failing ci linting
// due to importing enterprise path in oss folders.
export const AdminSidebarItems: Story = () => (
    <WebStory>
        {webProps => (
            <Grid columnCount={5}>
                <Code>isSourcegraphApp=true</Code>
                <Code>default</Code>
                <Code>isSourcegraphDotCom=true</Code>
                <Code>batchChangesEnabled=false</Code>
                <Code>codeInsightsEnabled=false</Code>
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={false}
                    isSourcegraphApp={true}
                    batchChangesEnabled={true}
                    batchChangesExecutionEnabled={true}
                    batchChangesWebhookLogsEnabled={true}
                    codeInsightsEnabled={true}
                    endUserOnboardingEnabled={false}
                />
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={false}
                    isSourcegraphApp={false}
                    batchChangesEnabled={true}
                    batchChangesExecutionEnabled={true}
                    batchChangesWebhookLogsEnabled={true}
                    codeInsightsEnabled={true}
                    endUserOnboardingEnabled={false}
                />
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={true}
                    isSourcegraphApp={false}
                    batchChangesEnabled={true}
                    batchChangesExecutionEnabled={true}
                    batchChangesWebhookLogsEnabled={true}
                    codeInsightsEnabled={true}
                    endUserOnboardingEnabled={false}
                />
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={false}
                    isSourcegraphApp={false}
                    batchChangesEnabled={false}
                    batchChangesExecutionEnabled={false}
                    batchChangesWebhookLogsEnabled={false}
                    codeInsightsEnabled={true}
                    endUserOnboardingEnabled={false}
                />
                <SiteAdminSidebar
                    {...webProps}
                    groups={enterpriseSiteAdminSidebarGroups}
                    isSourcegraphDotCom={false}
                    isSourcegraphApp={false}
                    batchChangesEnabled={true}
                    batchChangesExecutionEnabled={true}
                    batchChangesWebhookLogsEnabled={true}
                    codeInsightsEnabled={false}
                    endUserOnboardingEnabled={false}
                />
            </Grid>
        )}
    </WebStory>
)

AdminSidebarItems.storyName = 'Admin Sidebar Items'
AdminSidebarItems.parameters = {
    design: {
        type: 'figma',
        url: 'https://www.figma.com/file/cT23UcGqbdjdV4H2yxweIu/%2311775-Map-the-current-information-architecture-%5BApproved%5D?node-id=68%3A1',
    },
}
