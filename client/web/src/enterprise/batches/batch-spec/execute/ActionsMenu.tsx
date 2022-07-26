import React, { useCallback, useState } from 'react'

import { mdiChevronDown, mdiAlertCircle, mdiCloseCircleOutline, mdiPencil, mdiClose, mdiSync } from '@mdi/js'
import { VisuallyHidden } from '@reach/visually-hidden'
import { noop } from 'lodash'
import { useHistory, useLocation } from 'react-router'

import { useMutation } from '@sourcegraph/http-client'
import {
    Button,
    ButtonGroup,
    Icon,
    Link,
    Menu,
    MenuButton,
    MenuItem,
    MenuLink,
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

    // The Preview dropdown button is wider than the Actions menu,
    // so to prevent layout shift, we apply the width of the Preview
    // dropdown button to the Action button instead.
    const [menuReference, { width: menuWidth }] = useMeasure()

    return (
        <div className="position-relative">
            {showPreviewButton ? (
                // eslint-disable-next-line react/forbid-dom-props
                <div style={{ width: menuWidth }} className={styles.menuButton}>
                    <Menu>
                        <ButtonGroup>
                            <Button to={`${batchSpec.executionURL}/preview`} variant="primary" as={Link}>
                                Preview
                            </Button>
                            <MenuButton variant="primary" className={styles.dropdownButton}>
                                <Icon aria-hidden={true} svgPath={mdiChevronDown} />
                                <VisuallyHidden>Actions</VisuallyHidden>
                            </MenuButton>
                        </ButtonGroup>
                        <MenuList position={Position.bottomEnd}>
                            <MenuLink as={Link} to={`${batchChange.url}/close`}>
                                <Icon
                                    aria-hidden={true}
                                    className={styles.cancelIcon}
                                    svgPath={mdiCloseCircleOutline}
                                />{' '}
                                Close batch change
                            </MenuLink>
                        </MenuList>
                    </Menu>
                </div>
            ) : (
                <>
                    <div className={styles.menuButton}>
                        <Menu>
                            <div className="d-inline-block" aria-hidden={showPreviewButton}>
                                <MenuButton
                                    variant="secondary"
                                    style={{ width: menuWidth }}
                                    className={styles.actionsButton}
                                >
                                    <div className={styles.actionLabel}>Actions</div>
                                    <div className={styles.actionChevron}>
                                        <Icon
                                            aria-hidden={true}
                                            className={styles.chevronIcon}
                                            svgPath={mdiChevronDown}
                                        />
                                    </div>
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
                                        <Icon aria-hidden={true} className={styles.cancelIcon} svgPath={mdiClose} />{' '}
                                        Cancel execution...
                                    </MenuItem>
                                )}
                                {batchSpec.viewerCanRetry && (
                                    <MenuItem onSelect={retryBatchSpecExecution} disabled={isRetryLoading}>
                                        <Icon aria-hidden={true} svgPath={mdiSync} /> Retry failed workspaces
                                    </MenuItem>
                                )}
                                <MenuLink as={Link} to={`${batchChange.url}/close`}>
                                    <Icon
                                        aria-hidden={true}
                                        className={styles.cancelIcon}
                                        svgPath={mdiCloseCircleOutline}
                                    />{' '}
                                    Close batch change
                                </MenuLink>
                            </MenuList>
                        </Menu>
                    </div>
                    <CancelExecutionModal
                        isOpen={showCancelModal}
                        onCancel={() => setShowCancelModal(false)}
                        onConfirm={cancelModalType === 'cancel' ? cancelBatchSpecExecution : cancelAndEdit}
                        modalHeader={
                            cancelModalType === 'cancel' ? 'Cancel execution' : 'The execution is still running'
                        }
                        modalBody={
                            <Text>
                                {cancelModalType === 'cancel'
                                    ? 'Are you sure you want to cancel the current execution?'
                                    : 'You are unable to edit the spec when an execution is running.'}
                            </Text>
                        }
                        isLoading={isCancelLoading}
                    />
                </>
            )}
            {/* We need to render a Preview dropdown button, but make it invisible and non-actionable. */}
            <div className={styles.menuButtonHidden} ref={menuReference}>
                <Menu>
                    <ButtonGroup>
                        <Button to={`${batchSpec.executionURL}/preview`} variant="primary" as={Link}>
                            Preview
                        </Button>
                        <MenuButton variant="primary" className={styles.dropdownButton}>
                            <Icon aria-hidden={true} svgPath={mdiChevronDown} />
                            <VisuallyHidden>Actions</VisuallyHidden>
                        </MenuButton>
                    </ButtonGroup>
                    <MenuList position={Position.bottomEnd}>
                        <MenuLink as={Link} to={`${batchChange.url}/close`}>
                            <Icon aria-hidden={true} className={styles.cancelIcon} svgPath={mdiCloseCircleOutline} />{' '}
                            Close batch change
                        </MenuLink>
                    </MenuList>
                </Menu>
            </div>
        </div>
    )
})
