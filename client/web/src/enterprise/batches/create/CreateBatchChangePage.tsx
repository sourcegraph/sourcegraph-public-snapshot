import React from 'react'

import { SettingsCascadeProps } from '@sourcegraph/client-api'
import { Settings } from '@sourcegraph/shared/src/schema/settings.schema'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageHeader, Link } from '@sourcegraph/wildcard'

import { isBatchChangesExecutionEnabled } from '../../../batches'
import { BatchChangesIcon } from '../../../batches/icons'
import { PageTitle } from '../../../components/PageTitle'
import { Scalars } from '../../../graphql-operations'

import { CreateOrEditBatchChangePage } from './CreateOrEditBatchChangePage'
import { OldBatchChangePageContent } from './OldCreateBatchChangeContent'

export interface CreateBatchChangePageProps extends SettingsCascadeProps<Settings>, ThemeProps {
    // TODO: This can go away once we only have the new SSBC create page
    headingElement: 'h1' | 'h2'
    /**
     * The id for the namespace that the batch change should be created in, or that it
     * already belongs to, if it already exists.
     */
    initialNamespaceID?: Scalars['ID']
}

/**
 * CreateBatchChangePage is a wrapper around the create/edit batch change page that
 * determines if we should display the original create page or the new SSBC page.
 */
export const CreateBatchChangePage: React.FunctionComponent<CreateBatchChangePageProps> = ({
    settingsCascade,
    isLightTheme,
    headingElement,
    initialNamespaceID,
}) =>
    isBatchChangesExecutionEnabled(settingsCascade) ? (
        <CreateOrEditBatchChangePage
            isLightTheme={isLightTheme}
            settingsCascade={settingsCascade}
            initialNamespaceID={initialNamespaceID}
        />
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
