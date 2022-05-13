import React from 'react'

import FileDocumentIcon from 'mdi-react/FileDocumentIcon'
import SourceBranchIcon from 'mdi-react/SourceBranchIcon'
import { useHistory, useLocation } from 'react-router'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Badge, Container, Icon } from '@sourcegraph/wildcard'

import { BatchSpecFields } from '../../../graphql-operations'
import {
    BatchChangeTab,
    BatchChangeTabList,
    BatchChangeTabPanel,
    BatchChangeTabPanels,
    BatchChangeTabs,
} from '../BatchChangeTabs'
import { BatchSpec, BatchSpecDownloadButton } from '../BatchSpec'

import { PreviewPageAuthenticatedUser } from './BatchChangePreviewPage'
import {
    queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs,
    queryChangesetApplyPreview as _queryChangesetApplyPreview,
} from './list/backend'
import { PreviewList } from './list/PreviewList'

import styles from './BatchChangePreviewTabs.module.scss'

export interface BatchChangePreviewProps extends ThemeProps, TelemetryProps {
    batchSpecID: string
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

export const BatchChangePreviewTabs: React.FunctionComponent<React.PropsWithChildren<BatchChangePreviewTabsProps>> = ({
    authenticatedUser,
    batchSpecID,
    expandChangesetDescriptions,
    isLightTheme,
    queryChangesetApplyPreview,
    queryChangesetSpecFileDiffs,
    spec,
}) => {
    const history = useHistory()
    const location = useLocation()
    return (
        <BatchChangeTabs history={history} location={location}>
            <BatchChangeTabList>
                <BatchChangeTab index={0} name="previewchangesets">
                    <span>
                        <Icon className="text-muted mr-1" as={SourceBranchIcon} />
                        <span className="text-content" data-tab-content="Preview changesets">
                            Preview changesets
                        </span>{' '}
                        <Badge variant="secondary" pill={true} className="ml-1">
                            {spec.applyPreview.totalCount}
                        </Badge>
                    </span>
                </BatchChangeTab>
                <BatchChangeTab index={1} name="spec">
                    <span>
                        <Icon className="text-muted mr-1" as={FileDocumentIcon} />{' '}
                        <span className="text-content" data-tab-content="Spec">
                            Spec
                        </span>
                    </span>
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
                    <div className="d-flex mb-2 justify-content-end">
                        <BatchSpecDownloadButton
                            name={spec.description.name}
                            originalInput={spec.originalInput}
                            isLightTheme={isLightTheme}
                        />
                    </div>
                    <Container>
                        <BatchSpec
                            name={spec.description.name}
                            originalInput={spec.originalInput}
                            isLightTheme={isLightTheme}
                            className={styles.batchSpec}
                        />
                    </Container>
                </BatchChangeTabPanel>
            </BatchChangeTabPanels>
        </BatchChangeTabs>
    )
}
