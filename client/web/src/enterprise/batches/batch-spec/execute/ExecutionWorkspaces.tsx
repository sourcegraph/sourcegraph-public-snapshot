import { useCallback, useContext } from 'react'

import { useHistory } from 'react-router'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Card, CardBody, Panel } from '@sourcegraph/wildcard'

import { BatchSpecContext } from '../BatchSpecContext'

import { WorkspaceDetails } from './WorkspaceDetails'
import { Workspaces } from './workspaces/Workspaces'

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
    const { batchSpec, errors } = useContext(BatchSpecContext)

    const deselectWorkspace = useCallback(() => history.push(batchSpec.executionURL), [batchSpec.executionURL, history])

    return (
        <>
            {errors.execute && <ErrorAlert error={errors.execute} />}
            <div className={styles.container}>
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
        </>
    )
}
