import React, { useCallback, useEffect, useRef, useState } from 'react'

import { VisuallyHidden } from '@reach/visually-hidden'
import classNames from 'classnames'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { isErrorLike } from '@sourcegraph/common'
import { Alert, Button, Card, H4 } from '@sourcegraph/wildcard'

import styles from '../CodeMonitorForm.module.scss'

export interface ActionEditorProps {
    title: React.ReactNode
    subtitle: string
    disabled: boolean
    completed: boolean
    completedSubtitle: string
    idName: string // Name used for generating IDs, including form control IDs and test IDs

    actionEnabled: boolean
    toggleActionEnabled: (enabled: boolean, saveImmediately: boolean) => void

    includeResults: boolean
    toggleIncludeResults: (includeResults: boolean) => void

    canSubmit?: boolean
    onSubmit: React.FormEventHandler
    onCancel?: React.FormEventHandler

    canDelete: boolean
    onDelete: React.FormEventHandler

    warningMessage?: React.ReactNode

    // Test action
    testState: 'called' | 'loading' | Error | undefined

    testButtonDisabledReason?: string // If defined, the test button is disabled and this is the reason why
    testButtonText: string
    testAgainButtonText: string
    onTest: () => void

    // For testing purposes only
    _testStartOpen?: boolean
}

export const ActionEditor: React.FunctionComponent<React.PropsWithChildren<ActionEditorProps>> = ({
    title,
    subtitle,
    disabled,
    completed,
    completedSubtitle,
    idName,
    actionEnabled,
    toggleActionEnabled,
    includeResults,
    toggleIncludeResults,
    canSubmit = true,
    onSubmit,
    onCancel,
    canDelete,
    onDelete,
    warningMessage,
    testState,
    testButtonDisabledReason,
    testButtonText,
    testAgainButtonText,
    onTest,
    children,
    _testStartOpen = false,
}) => {
    const [expanded, setExpanded] = useState(_testStartOpen)

    // Focus card (or edit button) when collapsing the card.
    // Since the card starts collapsed, don't do it on the first render.
    const isFirstRender = useRef(true)
    const collapsedCard = useRef<HTMLDivElement & HTMLButtonElement>(null)
    const editLink = useRef<HTMLButtonElement>(null)
    useEffect((): void => {
        if (!isFirstRender.current && !expanded) {
            if (completed) {
                editLink.current?.focus()
            } else {
                collapsedCard.current?.focus()
            }
        }
        isFirstRender.current = false

        // Only focus when card is collapsed, not when anything else changes.
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [expanded])

    const submitHandler: React.FormEventHandler = useCallback(
        event => {
            setExpanded(false)
            onSubmit(event)
        },
        [onSubmit]
    )

    const cancelHandler: React.FormEventHandler = useCallback(
        event => {
            setExpanded(false)
            if (onCancel) {
                onCancel(event)
            }
        },
        [onCancel]
    )

    const deleteHandler: React.FormEventHandler = useCallback(
        event => {
            setExpanded(false)
            onDelete(event)
        },
        [onDelete]
    )

    return (
        <>
            {expanded && (
                <Card className={classNames(styles.card, 'p-3')}>
                    <H4 className="mb-0 font-weight-bold">{title}</H4>
                    <span className="text-muted">{subtitle}</span>

                    {warningMessage && (
                        <Alert variant="warning" className={classNames('mb-0 mt-3', styles.alert)}>
                            {warningMessage}
                        </Alert>
                    )}

                    {children}

                    <div className="d-flex align-items-center mb-3">
                        <div>
                            <Toggle
                                value={includeResults}
                                onToggle={toggleIncludeResults}
                                className="mr-2"
                                aria-labelledby={`code-monitoring-${idName}-include-results-toggle`}
                                data-testid={`include-results-toggle-${idName}`}
                            />
                        </div>
                        <span id={`code-monitoring-${idName}-include-results-toggle`}>
                            Include search results in sent message
                        </span>
                    </div>

                    <div className="flex mt-1">
                        <Button
                            className="mr-2"
                            variant="secondary"
                            outline={!testButtonDisabledReason}
                            disabled={!!testButtonDisabledReason || testState === 'loading' || testState === 'called'}
                            onClick={onTest}
                            size="sm"
                            data-testid={`send-test-${idName}`}
                        >
                            {testButtonText}
                        </Button>
                        {testState === 'called' && !testButtonDisabledReason && (
                            <Button
                                className="p-0"
                                onClick={onTest}
                                variant="link"
                                size="sm"
                                data-testid={`send-test-${idName}-again`}
                            >
                                {testAgainButtonText}
                            </Button>
                        )}

                        {testButtonDisabledReason && (
                            <div aria-live="polite" className={classNames('mt-2', styles.testActionError)}>
                                {testButtonDisabledReason}
                            </div>
                        )}

                        {isErrorLike(testState) && (
                            <div
                                aria-live="polite"
                                className={classNames('mt-2', styles.testActionError)}
                                data-testid={`test-${idName}-error`}
                            >
                                {testState.message}
                            </div>
                        )}
                    </div>

                    <div className="d-flex align-items-center my-4">
                        <div>
                            <Toggle
                                value={actionEnabled}
                                onToggle={enabled => toggleActionEnabled(enabled, !expanded)}
                                className="mr-2"
                                aria-labelledby={`code-monitoring-${idName}-form-actions-enable-toggle`}
                                data-testid={`enable-action-toggle-expanded-${idName}`}
                            />
                        </div>
                        <span id={`code-monitoring-${idName}-form-actions-enable-toggle`}>
                            {actionEnabled ? 'Enabled' : 'Disabled'}
                        </span>
                    </div>
                    <div className={styles.actionButtonRow}>
                        <Button
                            data-testid={`submit-action-${idName}`}
                            className={`test-submit-action-${idName}`}
                            onClick={submitHandler}
                            disabled={!canSubmit}
                            variant="secondary"
                        >
                            Continue
                        </Button>
                        <Button
                            onClick={cancelHandler}
                            outline={true}
                            variant="secondary"
                            data-testid={`cancel-action-${idName}`}
                        >
                            Cancel
                        </Button>
                        {canDelete && (
                            <>
                                <div className="flex-grow-1" />
                                <Button
                                    onClick={deleteHandler}
                                    outline={true}
                                    variant="danger"
                                    data-testid={`delete-action-${idName}`}
                                >
                                    Delete
                                </Button>
                            </>
                        )}
                    </div>
                </Card>
            )}
            {!expanded && (
                <Card
                    // When the action is completed, the wrapper cannot be a button because we show nested buttons inside it.
                    // Use a div instead. The edit button will still allow keyboard users to activate the form.
                    as={completed ? 'div' : Button}
                    data-testid={`form-action-toggle-${idName}`}
                    className={classNames(
                        `test-action-button-${idName}`,
                        styles.cardButton,
                        disabled && styles.btnDisabled
                    )}
                    disabled={disabled}
                    onClick={() => setExpanded(true)}
                    ref={collapsedCard}
                >
                    <div className="d-flex flex-wrap justify-content-between align-items-center w-100">
                        <div>
                            <VisuallyHidden>Edit action: </VisuallyHidden>
                            <div className={classNames('font-weight-bold', !completed && styles.cardLink)}>{title}</div>
                            {completed ? (
                                <span
                                    className="text-muted font-weight-normal"
                                    data-testid={`existing-action-${idName}`}
                                >
                                    {completedSubtitle}
                                </span>
                            ) : (
                                <span className="text-muted font-weight-normal">{subtitle}</span>
                            )}
                        </div>
                        {completed && (
                            <div className="d-flex align-items-center">
                                <div>
                                    <Toggle
                                        title={actionEnabled ? 'Enabled' : 'Disabled'}
                                        value={actionEnabled}
                                        onToggle={enabled => toggleActionEnabled(enabled, !expanded)}
                                        className="mr-3"
                                        data-testid={`enable-action-toggle-collapsed-${idName}`}
                                    />
                                </div>
                                <Button variant="link" className="p-0" ref={editLink}>
                                    Edit
                                </Button>
                            </div>
                        )}
                    </div>
                    {warningMessage && (
                        <Alert variant="warning" className={classNames('mb-0 mt-3', styles.alert)}>
                            {warningMessage}
                        </Alert>
                    )}
                </Card>
            )}
        </>
    )
}
