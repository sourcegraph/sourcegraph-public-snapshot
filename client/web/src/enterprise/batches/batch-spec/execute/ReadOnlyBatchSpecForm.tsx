import React, { useMemo, useState } from 'react'

import { useHistory } from 'react-router'

import { useMutation } from '@sourcegraph/http-client'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import { Button, H3, H4, Link, Text } from '@sourcegraph/wildcard'

import {
    BatchSpecExecutionFields,
    BatchSpecSource,
    CancelBatchSpecExecutionResult,
    CancelBatchSpecExecutionVariables,
} from '../../../../graphql-operations'
import { BatchSpec } from '../../BatchSpec'
import { BatchSpecContextState, useBatchSpecContext } from '../BatchSpecContext'
import { LibraryPane } from '../edit/library/LibraryPane'
import { WorkspacesPreviewPanel } from '../edit/workspaces-preview/WorkspacesPreviewPanel'

import { CANCEL_BATCH_SPEC_EXECUTION } from './backend'
import { CancelExecutionModal } from './CancelExecutionModal'
import { ReadOnlyBatchSpecAlert } from './ReadOnlyBatchSpecAlert'

import editorStyles from '../edit/EditBatchSpecPage.module.scss'

interface ReadOnlyBatchSpecFormProps extends ThemeProps {}

export const ReadOnlyBatchSpecForm: React.FunctionComponent<React.PropsWithChildren<ReadOnlyBatchSpecFormProps>> = ({
    isLightTheme,
}) => {
    const { batchChange, batchSpec, setActionsError } = useBatchSpecContext<BatchSpecExecutionFields>()

    return (
        <MemoizedReadOnlyBatchSpecForm
            isLightTheme={isLightTheme}
            batchChange={batchChange}
            batchSpec={batchSpec}
            setActionsError={setActionsError}
        />
    )
}

type MemoizedReadOnlyBatchSpecFormProps = ReadOnlyBatchSpecFormProps &
    Pick<BatchSpecContextState, 'batchChange' | 'batchSpec' | 'setActionsError'>

const MemoizedReadOnlyBatchSpecForm: React.FunctionComponent<
    React.PropsWithChildren<MemoizedReadOnlyBatchSpecFormProps>
> = React.memo(function MemoizedReadOnlyBatchSpecForm({ isLightTheme, batchChange, batchSpec, setActionsError }) {
    const history = useHistory()

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

    const alert: JSX.Element = useMemo(() => {
        if (batchSpec.source === BatchSpecSource.LOCAL) {
            return (
                <ReadOnlyBatchSpecAlert
                    variant="info"
                    className="d-flex align-items-center pr-3"
                    header="This spec is read-only"
                    message={
                        <>
                            This spec is read-only because it was created and executed locally with the{' '}
                            <Link to="/help/cli">Sourcegraph CLI (src)</Link>.
                        </>
                    }
                >
                    <Button variant="primary" onClick={() => history.push(`${batchChange.url}/edit`)}>
                        Edit spec
                    </Button>
                </ReadOnlyBatchSpecAlert>
            )
        }
        if (batchSpec.isExecuting) {
            return (
                <ReadOnlyBatchSpecAlert
                    variant="warning"
                    className="d-flex align-items-center pr-3"
                    header="The execution is still running"
                    message="You are unable to edit the spec when an execution is running."
                >
                    <Button variant="danger" onClick={() => setShowCancelModal(true)}>
                        Cancel execution
                    </Button>
                </ReadOnlyBatchSpecAlert>
            )
        }
        return (
            <ReadOnlyBatchSpecAlert
                variant="info"
                className="d-flex align-items-center pr-3"
                header="This spec is read-only"
                message="We've preserved the original batch spec from this execution for you to inspect."
            >
                {/* NOTE: As a future design consideration, what does the workflow look like if we
                open this in a new tab to allow the user to continue to reference their original
                batch spec at the same time? */}
                <Button variant="primary" onClick={() => history.push(`${batchChange.url}/edit`)}>
                    Edit spec
                </Button>
            </ReadOnlyBatchSpecAlert>
        )
    }, [batchSpec.isExecuting, batchChange.url, batchSpec.source, history])

    return (
        <div className={editorStyles.form}>
            <LibraryPane name={batchChange.name} isReadOnly={true} />
            <div className={editorStyles.editorContainer}>
                <H4 as={H3} className={editorStyles.header}>
                    Batch spec
                </H4>
                {alert}
                <BatchSpec
                    name={batchChange.name}
                    className={editorStyles.editor}
                    isLightTheme={isLightTheme}
                    originalInput={batchSpec.originalInput}
                />
            </div>
            {/* Hide the workspaces preview panel for locally-executed batch specs. */}
            {batchSpec.source === BatchSpecSource.REMOTE && <WorkspacesPreviewPanel isReadOnly={true} />}
            <CancelExecutionModal
                isOpen={showCancelModal}
                onCancel={() => setShowCancelModal(false)}
                onConfirm={cancelBatchSpecExecution}
                modalBody={<Text>Are you sure you want to cancel the current execution?</Text>}
                isLoading={isCancelLoading}
            />
        </div>
    )
})
