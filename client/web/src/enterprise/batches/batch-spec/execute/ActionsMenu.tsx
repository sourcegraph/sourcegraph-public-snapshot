import React, { useCallback, useState } from 'react'

import { mdiChevronDown, mdiAlertCircle, mdiPencil, mdiClose, mdiSync } from '@mdi/js'
import { noop } from 'lodash'
import { useHistory, useLocation } from 'react-router'

import { useMutation } from '@sourcegraph/http-client'
import {
    Button,
    Icon,
    Link,
    Menu,
    MenuButton,
    MenuItem,
    MenuList,
    Position,
    Text,
    useMeasure,
} from '@sourcegraph/wildcard'

import {
    BatchSpecExecutionFields,
    BatchSpecState,
    CancelBatchSpecExecutionResult,
    CancelBatchSpecExecutionVariables,
    RetryBatchSpecExecutionResult,
    RetryBatchSpecExecutionVariables,
} from '../../../../graphql-operations'
import { BatchSpecContextState, useBatchSpecContext } from '../BatchSpecContext'

import { CANCEL_BATCH_SPEC_EXECUTION, RETRY_BATCH_SPEC_EXECUTION } from './backend'
import { CancelExecutionModal } from './CancelExecutionModal'

import styles from './ActionsMenu.module.scss'

export const ActionsMenu: React.FunctionComponent = () => {
    const { batchChange, batchSpec, setActionsError } = useBatchSpecContext<BatchSpecExecutionFields>()

    return <MemoizedActionsMenu batchChange={batchChange} batchSpec={batchSpec} setActionsError={setActionsError} />
}

const MemoizedActionsMenu: React.FunctionComponent<
    React.PropsWithChildren<Pick<BatchSpecContextState, 'batchChange' | 'batchSpec' | 'setActionsError'>>
> = React.memo(function MemoizedActionsMenu({ batchChange, batchSpec, setActionsError }) {
    const history = useHistory()
    const location = useLocation()

    const { url } = batchChange
    const { isExecuting, state } = batchSpec

    const [showCancelModal, setShowCancelModal] = useState(false)
    const [cancelModalType, setCancelModalType] = useState<'cancel' | 'edit'>('cancel')
    const [cancelBatchSpecExecution, { loading: isCancelLoading }] = useMutation<
        CancelBatchSpecExecutionResult,
        CancelBatchSpecExecutionVariables
    >(CANCEL_BATCH_SPEC_EXECUTION, {
        variables: { id: batchSpec.id },
        onError: setActionsError,
        onCompleted: () => setShowCancelModal(false),
    })

    const cancelAndEdit = useCallback(() => {
        cancelBatchSpecExecution()
            .then(() => history.push(`${url}/edit`))
            .catch(noop)
    }, [cancelBatchSpecExecution, history, url])

    const [retryBatchSpecExecution, { loading: isRetryLoading }] = useMutation<
        RetryBatchSpecExecutionResult,
        RetryBatchSpecExecutionVariables
    >(RETRY_BATCH_SPEC_EXECUTION, { variables: { id: batchSpec.id }, onError: setActionsError })

    const onSelectEdit = useCallback(() => {
        if (isExecuting) {
            setCancelModalType('edit')
            setShowCancelModal(true)
        } else {
            history.push(`${url}/edit`)
        }
    }, [isExecuting, url, history])

    const onSelectCancel = useCallback(() => {
        setCancelModalType('cancel')
        setShowCancelModal(true)
    }, [])

    const showPreviewButton = !location.pathname.endsWith('preview') && state === BatchSpecState.COMPLETED
    const showPreviewMenuItem = !location.pathname.endsWith('preview') && state === BatchSpecState.FAILED

    // The actions menu button is wider than the "Preview" button, so to prevent layout
    // shift, we apply the width of the actions menu button to the "Preview" button
    // instead.
    const [menuReference, { width: menuWidth }] = useMeasure()

    return (
        <div className="position-relative">
            {showPreviewButton && (
                <Button
                    to={`${batchSpec.executionURL}/preview`}
                    variant="primary"
                    as={Link}
                    className={styles.previewButton}
                    style={{ width: menuWidth }}
                >
                    Preview
                </Button>
            )}
            <Menu>
                <div className="d-inline-block" ref={menuReference} aria-hidden={showPreviewButton}>
                    <MenuButton
                        variant="secondary"
                        className={showPreviewButton ? styles.menuButtonHidden : undefined}
                        // If an element with aria-hidden={true} contains a focusable
                        // element, assistive technologies won't read the focusable
                        // element, but keyboard users will still be able to navigate to
                        // it, which can cause confusion. We pair this with negative tab
                        // index to take the menu button out of the tab order when it is
                        // hidden. See: https://web.dev/aria-hidden-focus/
                        tabIndex={showPreviewButton ? -1 : undefined}
                    >
                        Actions
                        <Icon aria-hidden={true} className={styles.chevronIcon} svgPath={mdiChevronDown} />
                    </MenuButton>
                </div>
                <MenuList position={Position.bottomEnd}>
                    {showPreviewMenuItem && (
                        <MenuItem onSelect={() => history.push(`${batchSpec.executionURL}/preview`)}>
                            <Icon aria-hidden={true} svgPath={mdiAlertCircle} /> Preview with errors
                        </MenuItem>
                    )}
                    <MenuItem onSelect={onSelectEdit}>
                        <Icon aria-hidden={true} svgPath={mdiPencil} /> Edit spec{isExecuting ? '...' : ''}
                    </MenuItem>
                    {isExecuting && (
                        <MenuItem onSelect={onSelectCancel}>
                            <Icon aria-hidden={true} className={styles.cancelIcon} svgPath={mdiClose} /> Cancel
                            execution...
                        </MenuItem>
                    )}
                    {batchSpec.viewerCanRetry && (
                        <MenuItem onSelect={retryBatchSpecExecution} disabled={isRetryLoading}>
                            <Icon aria-hidden={true} svgPath={mdiSync} /> Retry failed workspaces
                        </MenuItem>
                    )}
                </MenuList>
            </Menu>
            <CancelExecutionModal
                isOpen={showCancelModal}
                onCancel={() => setShowCancelModal(false)}
                onConfirm={cancelModalType === 'cancel' ? cancelBatchSpecExecution : cancelAndEdit}
                modalHeader={cancelModalType === 'cancel' ? 'Cancel execution' : 'The execution is still running'}
                modalBody={
                    <Text>
                        {cancelModalType === 'cancel'
                            ? 'Are you sure you want to cancel the current execution?'
                            : 'You are unable to edit the spec when an execution is running.'}
                    </Text>
                }
                isLoading={isCancelLoading}
            />
        </div>
    )
})
