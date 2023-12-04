import React, { useState, useCallback, useMemo } from 'react'

import type * as H from 'history'
import { type Observable, concat, of } from 'rxjs'
import { switchMap, catchError, startWith, takeUntil, tap, delay } from 'rxjs/operators'

import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { type ErrorLike, isErrorLike, asError } from '@sourcegraph/common'
import { Button, LoadingSpinner, useEventObservable, Link, Alert } from '@sourcegraph/wildcard'

import type { CodeMonitorFields, ToggleCodeMonitorEnabledResult } from '../../graphql-operations'

import { toggleCodeMonitorEnabled as _toggleCodeMonitorEnabled } from './backend'

import styles from './CodeMonitoringNode.module.scss'

export interface CodeMonitorNodeProps {
    node: CodeMonitorFields
    location: H.Location
    showOwner: boolean

    toggleCodeMonitorEnabled?: typeof _toggleCodeMonitorEnabled
}

const LOADING = 'LOADING' as const

export const CodeMonitorNode: React.FunctionComponent<React.PropsWithChildren<CodeMonitorNodeProps>> = ({
    location,
    node,
    showOwner,
    toggleCodeMonitorEnabled = _toggleCodeMonitorEnabled,
}: CodeMonitorNodeProps) => {
    const [enabled, setEnabled] = useState<boolean>(node.enabled)

    const [toggleMonitor, toggleMonitorOrError] = useEventObservable(
        useCallback(
            (click: Observable<React.MouseEvent>) =>
                click.pipe(
                    switchMap(() => {
                        const toggleMonitor = toggleCodeMonitorEnabled(node.id, !enabled).pipe(
                            tap(
                                (
                                    idAndEnabled:
                                        | typeof LOADING
                                        | ErrorLike
                                        | ToggleCodeMonitorEnabledResult['toggleCodeMonitor']
                                ) => {
                                    if (idAndEnabled !== LOADING && !isErrorLike(idAndEnabled)) {
                                        setEnabled(idAndEnabled.enabled)
                                    }
                                }
                            ),
                            catchError(error => [asError(error)])
                        )
                        return concat(
                            of(LOADING).pipe(startWith(enabled), delay(300), takeUntil(toggleMonitor)),
                            toggleMonitor
                        )
                    })
                ),
            [node, enabled, setEnabled, toggleCodeMonitorEnabled]
        )
    )

    const actions = useMemo(
        () =>
            node.actions.nodes
                .map(action => {
                    switch (action.__typename) {
                        case 'MonitorEmail': {
                            return 'Sends email notification'
                        }
                        case 'MonitorSlackWebhook': {
                            return 'Sends Slack notification'
                        }
                        case 'MonitorWebhook': {
                            return 'Calls webhook'
                        }
                        default: {
                            return ''
                        }
                    }
                })
                .filter(string => string !== '')
                .join('; '),
        [node.actions]
    )

    return (
        <div className={styles.codeMonitoringNode}>
            <div className="d-flex justify-content-between align-items-center">
                <div className="d-flex flex-column">
                    <div className="font-weight-bold">
                        <Link to={`${location.pathname}/${node.id}`}>{node.description}</Link>
                        {showOwner && (
                            <>
                                {' '}
                                <Link
                                    className="text-muted"
                                    to={`${node.owner.url}/profile`}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                >
                                    (owned by {node.owner.namespaceName})
                                </Link>
                            </>
                        )}
                    </div>
                    {node.actions.nodes.length > 0 && (
                        <div className="d-flex text-muted align-items-center">New search result → {actions}</div>
                    )}
                </div>
                <div className="d-flex">
                    {toggleMonitorOrError === LOADING && <LoadingSpinner className="mr-2" />}
                    <div className={styles.toggleWrapper} data-testid="toggle-monitor-enabled">
                        <Toggle
                            onClick={toggleMonitor}
                            value={enabled}
                            className="mr-3"
                            disabled={toggleMonitorOrError === LOADING}
                            aria-label="Toggle code monitoring"
                        />
                    </div>
                    <Button
                        to={`${location.pathname}/${node.id}`}
                        className={styles.editButton}
                        variant="link"
                        as={Link}
                    >
                        Edit
                    </Button>
                </div>
            </div>
            {isErrorLike(toggleMonitorOrError) && (
                <Alert variant="danger">Failed to toggle monitor: {toggleMonitorOrError.message}</Alert>
            )}
        </div>
    )
}
