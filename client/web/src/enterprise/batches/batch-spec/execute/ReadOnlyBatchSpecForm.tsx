import React from 'react'

import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Alert, Button } from '@sourcegraph/wildcard'

import { BatchSpecState, EditBatchChangeFields } from '../../../../graphql-operations'
import { BatchSpec } from '../../BatchSpec'
import { LibraryPane } from '../edit/library/LibraryPane'
import { WorkspacesPreviewPanel } from '../edit/workspaces-preview/WorkspacesPreviewPanel'

import editorStyles from '../edit/EditBatchSpecPage.module.scss'

interface ReadOnlyBatchSpecFormProps extends ThemeProps {
    batchChange: EditBatchChangeFields
    originalInput: string
    executionState: BatchSpecState
}

export const ReadOnlyBatchSpecForm: React.FunctionComponent<React.PropsWithChildren<ReadOnlyBatchSpecFormProps>> = ({
    batchChange,
    originalInput,
    isLightTheme,
    executionState,
}) => {
    const alert =
        executionState === BatchSpecState.QUEUED || executionState === BatchSpecState.PROCESSING ? (
            <Alert variant="warning" className="d-flex align-items-center pr-3">
                <div className="flex-grow-1">
                    <h4>The execution is still running</h4>
                    You are unable to edit the spec when an execution is running.
                </div>
                {/* TODO: Handle button */}
                <Button variant="danger">Cancel execution</Button>
            </Alert>
        ) : (
            <Alert variant="info" className="d-flex align-items-center pr-3">
                <div className="flex-grow-1">
                    <h4>This spec is read-only</h4>
                    We've preserved the original batch spec from this execution for you to inspect.
                </div>
                <Button
                    variant="primary"
                    as="a"
                    href={`${batchChange.url}/edit`}
                    target="_blank"
                    rel="noopener noreferrer"
                >
                    Edit spec
                </Button>
            </Alert>
        )

    return (
        <div className={editorStyles.form}>
            <LibraryPane name={batchChange.name} isReadOnly={true} />
            <div className={editorStyles.editorContainer}>
                <h4 className={editorStyles.header}>Batch spec</h4>
                {alert}
                <BatchSpec
                    name={batchChange.name}
                    className={editorStyles.editor}
                    isLightTheme={isLightTheme}
                    originalInput={originalInput}
                />
            </div>
            <WorkspacesPreviewPanel isReadOnly={true} />
        </div>
    )
}
