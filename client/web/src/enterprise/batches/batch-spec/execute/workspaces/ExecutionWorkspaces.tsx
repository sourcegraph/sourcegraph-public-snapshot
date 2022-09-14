import React, { useCallback, useEffect, useRef } from 'react'

import { mdiClose } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import { useHistory } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { BatchSpecSource } from '@sourcegraph/shared/src/schema'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Card, CardBody, H3, H1, Icon, Text, Code } from '@sourcegraph/wildcard'

import { BatchSpecExecutionFields } from '../../../../../graphql-operations'
import { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from '../../../preview/list/backend'
import { BatchSpecContextState, useBatchSpecContext } from '../../BatchSpecContext'
import { queryBatchSpecWorkspaceStepFileDiffs as _queryBatchSpecWorkspaceStepFileDiffs } from '../backend'

import { WorkspaceDetails } from './WorkspaceDetails'
import { WorkspacesPanel } from './WorkspacesPanel'

import styles from './ExecutionWorkspaces.module.scss'

interface ExecutionWorkspacesProps extends ThemeProps {
    selectedWorkspaceID?: string
    /** For testing purposes only */
    queryBatchSpecWorkspaceStepFileDiffs?: typeof _queryBatchSpecWorkspaceStepFileDiffs
    queryChangesetSpecFileDiffs?: typeof _queryChangesetSpecFileDiffs
}

export const ExecutionWorkspaces: React.FunctionComponent<
    React.PropsWithChildren<ExecutionWorkspacesProps>
> = props => {
    const { batchSpec, errors } = useBatchSpecContext<BatchSpecExecutionFields>()

    if (batchSpec.source === BatchSpecSource.LOCAL) {
        return (
            <>
                <H1 className="text-center text-muted mt-5">
                    <Icon role="img" aria-hidden={true} svgPath={mdiClose} />
                    <VisuallyHidden>No Execution</VisuallyHidden>
                </H1>
                <Text alignment="center">
                    This batch spec was executed locally with <Code>src-cli</Code>.
                </Text>
            </>
        )
    }

    return <MemoizedExecutionWorkspaces {...props} batchSpec={batchSpec} errors={errors} />
}

type MemoizedExecutionWorkspacesProps = ExecutionWorkspacesProps & Pick<BatchSpecContextState, 'batchSpec' | 'errors'>

const MemoizedExecutionWorkspaces: React.FunctionComponent<
    React.PropsWithChildren<MemoizedExecutionWorkspacesProps>
> = React.memo(function MemoizedExecutionWorkspaces({
    selectedWorkspaceID,
    isLightTheme,
    batchSpec,
    errors,
    queryBatchSpecWorkspaceStepFileDiffs,
    queryChangesetSpecFileDiffs,
}) {
    const history = useHistory()

    const deselectWorkspace = useCallback(() => {
        history.push({ ...history.location, pathname: `${batchSpec.executionURL}/execution` })
    }, [batchSpec.executionURL, history])

    const videoRef = useRef<HTMLVideoElement | null>(null)
    // Pause the execution animation loop when the batch spec stops executing.
    useEffect(() => {
        if (!batchSpec.isExecuting) {
            videoRef.current?.pause()
        }
    }, [batchSpec.isExecuting])

    return (
        <div className={styles.container}>
            {errors.execute && <ErrorAlert error={errors.execute} className={styles.errors} />}
            <div className={styles.inner}>
                <WorkspacesPanel
                    batchSpecID={batchSpec.id}
                    selectedNode={selectedWorkspaceID}
                    executionURL={batchSpec.executionURL}
                />
                <Card className="w-100 overflow-auto flex-grow-1">
                    {/* This is necessary to prevent the margin collapse on `Card` */}
                    <div className="w-100">
                        <CardBody>
                            {selectedWorkspaceID ? (
                                <WorkspaceDetails
                                    id={selectedWorkspaceID}
                                    isLightTheme={isLightTheme}
                                    deselectWorkspace={deselectWorkspace}
                                    queryBatchSpecWorkspaceStepFileDiffs={queryBatchSpecWorkspaceStepFileDiffs}
                                    queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                                />
                            ) : (
                                <>
                                    <div className={styles.videoContainer}>
                                        <video
                                            className="w-100 percy-hide"
                                            autoPlay={true}
                                            muted={true}
                                            loop={true}
                                            playsInline={true}
                                            controls={false}
                                            ref={videoRef}
                                        >
                                            <source
                                                type="video/webm"
                                                src={`https://storage.googleapis.com/sourcegraph-assets/batch-changes/execution-animation${
                                                    isLightTheme ? '' : '-dark'
                                                }.webm`}
                                            />
                                            <source
                                                type="video/mp4"
                                                src={`https://storage.googleapis.com/sourcegraph-assets/batch-changes/execution-animation${
                                                    isLightTheme ? '' : '-dark'
                                                }.mp4`}
                                            />
                                        </video>
                                    </div>
                                    <H3 className="text-center my-3">Select a workspace to view details.</H3>
                                </>
                            )}
                        </CardBody>
                    </div>
                </Card>
            </div>
        </div>
    )
})
