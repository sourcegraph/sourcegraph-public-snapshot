import React, { useState } from 'react'

import { useHistory } from 'react-router'

import { useMutation } from '@sourcegraph/http-client'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Alert, Button } from '@sourcegraph/wildcard'

import {
    BatchSpecExecutionFields,
    BatchSpecState,
    CancelBatchSpecExecutionResult,
    CancelBatchSpecExecutionVariables,
} from '../../../../graphql-operations'
import { BatchSpec } from '../../BatchSpec'
import { useBatchSpecContext } from '../BatchSpecContext'
import { LibraryPane } from '../edit/library/LibraryPane'
import { WorkspacesPreviewPanel } from '../edit/workspaces-preview/WorkspacesPreviewPanel'

import { CANCEL_BATCH_SPEC_EXECUTION } from './backend'
import { CancelExecutionModal } from './CancelExecutionModal'

import editorStyles from '../edit/EditBatchSpecPage.module.scss'

interface ReadOnlyBatchSpecFormProps extends ThemeProps {}

export const ReadOnlyBatchSpecForm: React.FunctionComponent<React.PropsWithChildren<ReadOnlyBatchSpecFormProps>> = ({
    isLightTheme,
}) => {
    const history = useHistory()

    const { batchChange, batchSpec, setActionsError } = useBatchSpecContext<BatchSpecExecutionFields>()

    const [showCancelModal, setShowCancelModal] = useState(false)
    const [cancelBatchSpecExecution, { loading: isCancelLoading }] = useMutation<
        CancelBatchSpecExecutionResult,
        CancelBatchSpecExecutionVariables
    >(CANCEL_BATCH_SPEC_EXECUTION, {
        variables: { id: batchSpec.id },
        onError: setActionsError,
        onCompleted: () => {
            setShowCancelModal(false)
            history.push(`${batchChange.url}/edit`)
        },
    })

    const alert =
        batchSpec.state === BatchSpecState.QUEUED || batchSpec.state === BatchSpecState.PROCESSING ? (
            <Alert variant="warning" className="d-flex align-items-center pr-3">
                <div className="flex-grow-1">
                    <h4>The execution is still running</h4>
                    You are unable to edit the spec when an execution is running.
                </div>
                <Button variant="danger" onClick={() => setShowCancelModal(true)}>
                    Cancel execution
                </Button>
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
                    originalInput={batchSpec.originalInput}
                />
            </div>
            <WorkspacesPreviewPanel isReadOnly={true} />
            <CancelExecutionModal
                isOpen={showCancelModal}
                onCancel={() => setShowCancelModal(false)}
                onConfirm={cancelBatchSpecExecution}
                modalBody={<p>Are you sure you want to cancel the current execution?</p>}
                isLoading={isCancelLoading}
            />
        </div>
    )
}
