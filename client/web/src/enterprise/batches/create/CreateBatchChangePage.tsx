import React from 'react'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Page } from '@sourcegraph/web/src/components/Page'
import { PageHeader } from '@sourcegraph/wildcard'

import { isBatchChangesExecutionEnabled } from '../../../batches'
import { BatchChangesIcon } from '../../../batches/icons'
import { PageTitle } from '../../../components/PageTitle'
import { OrgAreaOrganizationFields, UserAreaUserFields } from '../../../graphql-operations'
import { Settings } from '../../../schema/settings.schema'

import { NewCreateBatchChangePage } from './NewCreateBatchChangePage'
import { OldBatchChangePageContent } from './OldCreateBatchChangeContent'

export interface CreateBatchChangePageProps extends SettingsCascadeProps<Settings>, ThemeProps {
    // TODO: This can go away once we only have the new SSBC create page
    headingElement: 'h1' | 'h2'
    namespace: UserAreaUserFields | OrgAreaOrganizationFields
    // TODO: This affects work on another branch
    settingsInitiallyOpen?: boolean
}

export const CreateBatchChangePage: React.FunctionComponent<CreateBatchChangePageProps> = ({
    settingsCascade,
    isLightTheme,
    headingElement,
    namespace,
    // TODO: This affects work on another branch
    settingsInitiallyOpen,
}) =>
    isBatchChangesExecutionEnabled(settingsCascade) ? (
        <NewCreateBatchChangePage
            isLightTheme={isLightTheme}
            settingsCascade={settingsCascade}
            namespace={namespace}
            settingsInitiallyOpen={settingsInitiallyOpen}
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
                        <a href="/help/batch_changes" rel="noopener noreferrer" target="_blank">
                            documentation.
                        </a>
                    </>
                }
                className="mb-3"
            />
            <OldBatchChangePageContent />
        </Page>
    )
