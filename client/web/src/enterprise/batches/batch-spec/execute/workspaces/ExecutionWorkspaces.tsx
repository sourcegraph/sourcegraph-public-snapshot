import { useCallback } from 'react'

import { useHistory } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Card, CardBody, Panel } from '@sourcegraph/wildcard'

import { BatchSpecExecutionFields } from '../../../../../graphql-operations'
import { useBatchSpecContext } from '../../BatchSpecContext'

import { WorkspaceDetails } from './WorkspaceDetails'
import { Workspaces } from './Workspaces'

import styles from './ExecutionWorkspaces.module.scss'

const WORKSPACES_LIST_SIZE = 'batch-changes.ssbc-workspaces-list-size'

interface ExecutionWorkspacesProps extends ThemeProps {
    selectedWorkspaceID?: string
}

export const ExecutionWorkspaces: React.FunctionComponent<React.PropsWithChildren<ExecutionWorkspacesProps>> = ({
    selectedWorkspaceID,
    isLightTheme,
}) => {
    const history = useHistory()
    const { batchSpec, errors } = useBatchSpecContext<BatchSpecExecutionFields>()

    const deselectWorkspace = useCallback(() => history.push(batchSpec.executionURL), [batchSpec.executionURL, history])

    return (
        <div className={styles.container}>
            {errors.execute && <ErrorAlert error={errors.execute} className={styles.errors} />}
            <div className={styles.inner}>
                <Panel defaultSize={500} minSize={405} maxSize={1400} position="left" storageKey={WORKSPACES_LIST_SIZE}>
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
                                />
                            ) : (
                                <h3 className="text-center my-3">Select a workspace to view details.</h3>
                            )}
                        </CardBody>
                    </div>
                </Card>
            </div>
        </div>
    )
}
