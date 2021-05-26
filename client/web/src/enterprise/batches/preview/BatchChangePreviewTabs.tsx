import * as H from 'history'
import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import React from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'

import { BatchSpecFields } from '../../../graphql-operations'
import {
    BatchChangeTab,
    BatchChangeTabList,
    BatchChangeTabPanel,
    BatchChangeTabPanels,
    BatchChangeTabs,
} from '../BatchChangeTabs'
import { BatchSpec, BatchSpecDownloadLink } from '../BatchSpec'

import { PreviewPageAuthenticatedUser } from './BatchChangePreviewPage'
import {
    queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs,
    queryChangesetApplyPreview as _queryChangesetApplyPreview,
} from './list/backend'
import { PreviewList } from './list/PreviewList'
import { Container } from '@sourcegraph/wildcard'

export interface BatchChangePreviewProps extends ThemeProps, TelemetryProps {
    batchSpecID: string
    history: H.History
    location: H.Location
    authenticatedUser: PreviewPageAuthenticatedUser

    /** Used for testing. */
    queryChangesetApplyPreview?: typeof _queryChangesetApplyPreview
    /** Used for testing. */
    queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. */
    expandChangesetDescriptions?: boolean
}

interface BatchChangePreviewTabsProps extends BatchChangePreviewProps {
    spec: BatchSpecFields
}

export const BatchChangePreviewTabs: React.FunctionComponent<BatchChangePreviewTabsProps> = ({
    authenticatedUser,
    batchSpecID,
    expandChangesetDescriptions,
    history,
    isLightTheme,
    location,
    queryChangesetApplyPreview,
    queryChangesetSpecFileDiffs,
    spec,
}) => (
    <BatchChangeTabs history={history} location={location}>
        <BatchChangeTabList>
            <BatchChangeTab index={0} name="previewchangesets">
                <SourceBranchIcon className="icon-inline text-muted mr-1" />
                Preview changesets{' '}
                <span className="badge badge-pill badge-secondary ml-1">{spec.applyPreview.totalCount}</span>
            </BatchChangeTab>
            <BatchChangeTab index={1} name="spec">
                <FileDocumentIcon className="icon-inline text-muted mr-1" /> Spec
            </BatchChangeTab>
        </BatchChangeTabList>
        <BatchChangeTabPanels>
            <BatchChangeTabPanel index={0}>
                <PreviewList
                    batchSpecID={batchSpecID}
                    history={history}
                    location={location}
                    authenticatedUser={authenticatedUser}
                    isLightTheme={isLightTheme}
                    queryChangesetApplyPreview={queryChangesetApplyPreview}
                    queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                    expandChangesetDescriptions={expandChangesetDescriptions}
                />
            </BatchChangeTabPanel>
            <BatchChangeTabPanel index={1}>
                <Container>
                    <div className="d-flex mb-2 justify-content-end">
                        <BatchSpecDownloadLink name={spec.description.name} originalInput={spec.originalInput} />
                    </div>
                    <BatchSpec originalInput={spec.originalInput} />
                </Container>
            </BatchChangeTabPanel>
        </BatchChangeTabPanels>
    </BatchChangeTabs>
)
