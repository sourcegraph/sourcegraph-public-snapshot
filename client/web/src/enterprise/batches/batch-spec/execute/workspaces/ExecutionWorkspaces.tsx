import React, { useCallback } from 'react'

import { useHistory } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Card, CardBody, Panel, H3 } from '@sourcegraph/wildcard'

import { BatchSpecExecutionFields } from '../../../../../graphql-operations'
import { queryChangesetSpecFileDiffs as _queryChangesetSpecFileDiffs } from '../../../preview/list/backend'
import { BatchSpecContextState, useBatchSpecContext } from '../../BatchSpecContext'
import { queryBatchSpecWorkspaceStepFileDiffs as _queryBatchSpecWorkspaceStepFileDiffs } from '../backend'

import { WorkspaceDetails } from './WorkspaceDetails'
import { Workspaces } from './Workspaces'

import styles from './ExecutionWorkspaces.module.scss'

const WORKSPACES_LIST_SIZE = 'batch-changes.ssbc-workspaces-list-size'

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

    return <MemoizedExecutionWorkspaces {...props} batchSpec={batchSpec} errors={errors} />
}

type MemoizedExecutionWorkspacesProps = ExecutionWorkspacesProps & Pick<BatchSpecContextState, 'batchSpec' | 'errors'>

const MemoizedExecutionWorkspaces: React.FunctionComponent<
    React.PropsWithChildren<MemoizedExecutionWorkspacesProps>
> = React.memo(
    ({
        selectedWorkspaceID,
        isLightTheme,
        batchSpec,
        errors,
        queryBatchSpecWorkspaceStepFileDiffs,
        queryChangesetSpecFileDiffs,
    }) => {
        const history = useHistory()

        const deselectWorkspace = useCallback(() => history.push(batchSpec.executionURL), [
            batchSpec.executionURL,
            history,
        ])

        return (
            <div className={styles.container}>
                {errors.execute && <ErrorAlert error={errors.execute} className={styles.errors} />}
                <div className={styles.inner}>
                    <Panel
                        defaultSize={500}
                        minSize={405}
                        maxSize={1400}
                        position="left"
                        storageKey={WORKSPACES_LIST_SIZE}
                    >
                        <Workspaces
                            batchSpecID={batchSpec.id}
                            selectedNode={selectedWorkspaceID}
                            executionURL={batchSpec.executionURL}
                        />
                    </Panel>
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
                                    <H3 className="text-center my-3">Select a workspace to view details.</H3>
                                )}
                            </CardBody>
                        </div>
                    </Card>
                </div>
            </div>
        )
    }
)
