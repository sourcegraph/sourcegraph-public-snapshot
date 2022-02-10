import classNames from 'classnames'
import SearchIcon from 'mdi-react/SearchIcon'
import WarningIcon from 'mdi-react/WarningIcon'
import React, { useEffect, useMemo, useState } from 'react'
import { animated, useSpring } from 'react-spring'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { UseConnectionResult } from '@sourcegraph/web/src/components/FilteredConnection/hooks/useConnection'
import { Button } from '@sourcegraph/wildcard'

import { Connection } from '../../../../components/FilteredConnection'
import { BatchSpecWorkspaceResolutionState, PreviewBatchSpecWorkspaceFields } from '../../../../graphql-operations'
import { ResolutionState } from '../useWorkspacesPreview'

import { ImportingChangesetsPreviewList } from './ImportingChangesetsPreviewList'
import { PreviewLoadingSpinner } from './PreviewLoadingSpinner'
import { PreviewPromptIcon } from './PreviewPromptIcon'
import { ImportingChangesetFields } from './useImportingChangesets'
import { WorkspacePreviewFilters } from './useWorkspaces'
import styles from './WorkspacesPreview.module.scss'
import { WorkspacePreviewFilterRow } from './WorkspacesPreviewFilterRow'
import { WorkspacesPreviewList } from './WorkspacesPreviewList'

/** Example snippet show in preview prompt if user has not yet added an on: statement. */
const ON_STATEMENT = `on:
  - repositoriesMatchingQuery: repo:my-org/.*
`

interface WorkspacesPreviewProps {
    /**
     * Function to submit the current input batch spec YAML to trigger a new workspaces
     * preview request.
     */
    preview: () => void
    /**
     * Whether or not the preview button should be disabled, for example due to there
     * being a problem with the input batch spec YAML, or a preview request already being
     * in flight. An optional tooltip string to display may be provided in place of
     * `true`.
     */
    previewDisabled: boolean | string
    /**
     * Whether or not the batch spec YAML on the server which was used to preview
     * workspaces is up-to-date with that which is presently in the editor.
     */
    batchSpecStale: boolean
    /** Whether or not the user has previewed their batch spec at least once. */
    hasPreviewed: boolean
    /**
     * Function to automatically update repo query of input batch spec YAML to exclude the
     * provided repo + branch.
     */
    excludeRepo: (repo: string, branch: string) => void
    /** Method to invoke to cancel any current workspaces resolution job. */
    cancel: () => void
    /**
     * Whether or not a preview request is currently in flight or the current workspaces
     * resolution job is in progress.
     */
    isWorkspacesPreviewInProgress: boolean
    /** The status of the current workspaces resolution job. */
    resolutionState: ResolutionState
    /** Any error from `previewBatchSpec` or the workspaces resolution job. */
    error?: string
    /** The current workspaces preview connection result used to render the list. */
    workspacesConnection: UseConnectionResult<PreviewBatchSpecWorkspaceFields>
    /** The current importing changesets connection result used to render the list. */
    importingChangesetsConnection: UseConnectionResult<ImportingChangesetFields>
    /** Method to invoke to capture a change in the active filters applied. */
    setFilters: (filters: WorkspacePreviewFilters) => void
}

export const WorkspacesPreview: React.FunctionComponent<WorkspacesPreviewProps> = ({
    previewDisabled,
    preview,
    batchSpecStale,
    hasPreviewed,
    excludeRepo,
    isWorkspacesPreviewInProgress,
    cancel,
    error,
    resolutionState,
    workspacesConnection,
    importingChangesetsConnection,
    setFilters,
}) => {
    const { connection } = workspacesConnection

    // Before we've ever previewed workspaces for this batch change, there's no reason to
    // show the list or filters for the connection.
    const shouldShowConnection = hasPreviewed || !!connection?.nodes.length

    // We "cache" the last results of the workspaces preview so that we can continue to
    // show them in the list while the next workspaces resolution is still in progress. We
    // have to do this outside of Apollo Client because we continue to requery the
    // workspaces preview while the resolution job is still in progress, and so the
    // results will come up empty and overwrite the previous results in the Apollo Client
    // cache while this is happening.
    const [cachedWorkspacesPreview, setCachedWorkspacesPreview] = useState<
        Connection<PreviewBatchSpecWorkspaceFields>
    >()

    // We copy the results from `connection` to `cachedWorkspacesPreview` whenever a
    // resolution job completes.
    useEffect(() => {
        if (resolutionState === BatchSpecWorkspaceResolutionState.COMPLETED && connection?.nodes.length) {
            setCachedWorkspacesPreview(connection)
        }
    }, [resolutionState, connection])

    // We will instruct `WorkspacesPreviewList` to show the cached results instead of
    // whatever is in `connection` if we know the workspaces preview resolution is
    // currently in progress.
    const showCached = useMemo(
        () =>
            Boolean(
                cachedWorkspacesPreview?.nodes.length &&
                    (isWorkspacesPreviewInProgress || resolutionState === 'CANCELED')
            ),
        [cachedWorkspacesPreview, isWorkspacesPreviewInProgress, resolutionState]
    )

    const ctaButton = isWorkspacesPreviewInProgress ? (
        <Button className="mt-3 mb-2" variant="secondary" onClick={cancel}>
            Cancel
        </Button>
    ) : (
        <Button
            className="mt-3 mb-2"
            variant="success"
            disabled={!!previewDisabled}
            data-tooltip={typeof previewDisabled === 'string' ? previewDisabled : undefined}
            onClick={preview}
        >
            <SearchIcon className="icon-inline mr-1" />
            {error ? 'Retry preview' : 'Preview workspaces'}
        </Button>
    )

    const [exampleOpen, setExampleOpen] = useState(false)
    const exampleStyle = useSpring({ height: exampleOpen ? '6.5rem' : '0rem', opacity: exampleOpen ? 1 : 0 })

    const ctaInstructions = isWorkspacesPreviewInProgress ? (
        <h4 className={styles.instruction}>Hang tight while we look for matching workspaces...</h4>
    ) : batchSpecStale ? (
        <h4 className={styles.instruction}>Finish editing your batch spec, then manually preview repositories.</h4>
    ) : (
        <>
            <h4 className={styles.instruction}>
                {hasPreviewed ? 'Modify your' : 'Add an'} <span className="text-monospace">on:</span> statement to
                preview repositories.
                <Button
                    className={styles.toggleExampleButton}
                    display="inline"
                    onClick={() => setExampleOpen(!exampleOpen)}
                >
                    {exampleOpen ? 'Close example' : 'See example'}
                </Button>
            </h4>
            <animated.div style={exampleStyle} className={styles.onExample}>
                <CodeSnippet className="w-100 mt-3" code={ON_STATEMENT} language="yaml" withCopyButton={true} />
            </animated.div>
        </>
    )

    return (
        <div className="d-flex flex-column align-items-center w-100 h-100">
            <h3 className={styles.header}>
                Workspaces preview{' '}
                {(batchSpecStale || !hasPreviewed) && shouldShowConnection && (
                    <WarningIcon
                        className="icon-inline text-muted"
                        data-tooltip="The workspaces previewed below may not be up-to-date."
                    />
                )}
            </h3>
            {/* We wrap this section in its own div to prevent margin collapsing within the flex column */}
            <div className="d-flex flex-column align-items-center w-100 mb-3">
                {error && <ErrorAlert error={error} className="w-100 mb-0" />}
                <div className={styles.iconContainer}>
                    <PreviewLoadingSpinner
                        className={classNames({ [styles.hidden]: !isWorkspacesPreviewInProgress })}
                    />
                    <PreviewPromptIcon className={classNames({ [styles.hidden]: isWorkspacesPreviewInProgress })} />
                </div>
                {ctaInstructions}
                {ctaButton}
            </div>
            {shouldShowConnection && (
                <WorkspacePreviewFilterRow onFiltersChange={setFilters} disabled={isWorkspacesPreviewInProgress} />
            )}
            {shouldShowConnection && (
                <div className="d-flex flex-column align-items-center overflow-auto w-100">
                    <WorkspacesPreviewList
                        isStale={batchSpecStale || !hasPreviewed}
                        excludeRepo={excludeRepo}
                        workspacesConnection={workspacesConnection}
                        showCached={showCached}
                        cached={cachedWorkspacesPreview}
                    />
                    <ImportingChangesetsPreviewList
                        isStale={batchSpecStale || !hasPreviewed}
                        importingChangesetsConnection={importingChangesetsConnection}
                    />
                </div>
            )}
        </div>
    )
}
