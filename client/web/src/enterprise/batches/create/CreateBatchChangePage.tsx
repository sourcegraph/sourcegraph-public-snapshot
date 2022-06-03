import React from 'react'

import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Link, PageHeader } from '@sourcegraph/wildcard'

import { isBatchChangesExecutionEnabled } from '../../../batches'
import { BatchChangesIcon } from '../../../batches/icons'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { Scalars } from '../../../graphql-operations'
import { BatchChangeHeader } from '../batch-spec/header/BatchChangeHeader'
import { TabBar, TabsConfig } from '../batch-spec/TabBar'

import { ConfigurationForm } from './ConfigurationForm'
import { InsightTemplatesBanner } from './InsightTemplatesBanner'
import { OldBatchChangePageContent } from './OldCreateBatchChangeContent'
import { useInsightTemplates } from './useInsightTemplates'

import layoutStyles from '../batch-spec/Layout.module.scss'

export interface CreateBatchChangePageProps extends SettingsCascadeProps<Settings>, ThemeProps {
    // TODO: This can go away once we only have the new SSBC create page
    headingElement: 'h1' | 'h2'
    initialNamespaceID?: Scalars['ID']
}

/**
 * CreateBatchChangePage is a wrapper around the create batch change page that determines
 * if we should display the original create page or the new server-side flow page.
 */
export const CreateBatchChangePage: React.FunctionComponent<React.PropsWithChildren<CreateBatchChangePageProps>> = ({
    settingsCascade,
    headingElement,
    ...props
}) =>
    isBatchChangesExecutionEnabled(settingsCascade) ? (
        <NewBatchChangePageContent settingsCascade={settingsCascade} {...props} />
    ) : (
        <Page>
            <PageTitle title="Create batch change" />
            <PageHeader
                path={[{ icon: BatchChangesIcon, text: 'Create batch change' }]}
                headingElement={headingElement}
                description={
                    <>
                        Follow these steps to create a Batch Change. Need help? View the{' '}
                        <Link to="/help/batch_changes" rel="noopener noreferrer" target="_blank">
                            documentation.
                        </Link>
                    </>
                }
                className="mb-3"
            />
            <OldBatchChangePageContent />
        </Page>
    )

const TABS_CONFIG: TabsConfig[] = [{ key: 'configuration', isEnabled: true }]

const NewBatchChangePageContent: React.FunctionComponent<
    React.PropsWithChildren<Omit<CreateBatchChangePageProps, 'headingElement'>>
> = ({ settingsCascade, initialNamespaceID }) => {
    const { renderTemplate, insightTitle } = useInsightTemplates(settingsCascade)
    return (
        <div className={layoutStyles.pageContainer}>
            <PageTitle title="Create new batch change" />
            {insightTitle && <InsightTemplatesBanner insightTitle={insightTitle} type="create" className="mb-5" />}
            <div className={layoutStyles.headerContainer}>
                <BatchChangeHeader title={{ text: 'Create batch change' }} />
            </div>
            <TabBar activeTabKey="configuration" tabsConfig={TABS_CONFIG} />
            <ConfigurationForm
                renderTemplate={renderTemplate}
                insightTitle={insightTitle}
                settingsCascade={settingsCascade}
                initialNamespaceID={initialNamespaceID}
            />
        </div>
    )
}
