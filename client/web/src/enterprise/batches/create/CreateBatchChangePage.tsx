import React from 'react'

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { PageHeader } from '@sourcegraph/wildcard'

import { isBatchChangesExecutionEnabled } from '../../../batches'
import { BatchChangesIcon } from '../../../batches/icons'
import { PageTitle } from '../../../components/PageTitle'
import { Settings } from '../../../schema/settings.schema'

import { createBatchSpecFromRaw as _createBatchSpecFromRaw } from './backend'
import { NewCreateBatchChangeContent } from './NewCreateBatchChangeContent'
import { OldBatchChangePageContent } from './OldCreateBatchChangeContent'

export interface CreateBatchChangePageProps extends SettingsCascadeProps<Settings>, ThemeProps {
    headingElement: 'h1' | 'h2'
    /* For testing only. */
    createBatchSpecFromRaw?: typeof _createBatchSpecFromRaw
}

export const CreateBatchChangePage: React.FunctionComponent<CreateBatchChangePageProps> = ({
    settingsCascade,
    isLightTheme,
    headingElement,
    createBatchSpecFromRaw,
}) => (
    <>
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
        {isBatchChangesExecutionEnabled(settingsCascade) ? (
            <NewCreateBatchChangeContent
                createBatchSpecFromRaw={createBatchSpecFromRaw}
                isLightTheme={isLightTheme}
                settingsCascade={settingsCascade}
            />
        ) : (
            <OldBatchChangePageContent />
        )}
    </>
)
