import React, { useCallback, useState } from 'react'

import { gql } from '@apollo/client'

import { useQuery } from '@sourcegraph/http-client'
import { Alert, ProductStatusBadge, Select } from '@sourcegraph/wildcard'

import {
    ListBatchChangeAutocompleteResult,
    ListBatchChangeAutocompleteVariables,
    Scalars,
} from '../../../../graphql-operations'
import { ActionProps } from '../FormActionArea'

import { ActionEditor } from './ActionEditor'

export const LIST_AUTOCOMPLETE = gql`
    query ListBatchChangeAutocomplete($namespace: ID!) {
        namespace(id: $namespace) {
            ... on User {
                batchChanges {
                    nodes {
                        id
                        namespace {
                            namespaceName
                        }
                        name
                        url
                    }
                }
            }
            ... on Org {
                batchChanges {
                    nodes {
                        id
                        namespace {
                            namespaceName
                        }
                        name
                        url
                    }
                }
            }
        }
    }
`

export const BatchChangeAction: React.FunctionComponent<ActionProps> = ({
    action,
    setAction,
    disabled,
    authenticatedUser,
    _testStartOpen,
}) => {
    const [isEnabled, setIsEnabled] = useState(action ? action.enabled : true)

    const toggleEnabled: (enabled: boolean, saveImmediately: boolean) => void = useCallback(
        (enabled, saveImmediately) => {
            setIsEnabled(enabled)
            if (action && saveImmediately) {
                setAction({ ...action, enabled })
            }
        },
        [action, setAction]
    )

    const [selectedBatchChange, setSelectedBatchChange] = useState<Scalars['ID']>()

    const onSubmit: React.FormEventHandler = useCallback(
        event => {
            event.preventDefault()
            setAction({
                __typename: 'MonitorBatchChange',
                id: action ? action.id : '',
                enabled: isEnabled,
                batchChange: selectedBatchChange,
            })
        },
        [action, setAction, isEnabled, selectedBatchChange]
    )

    const onCancel: React.FormEventHandler = useCallback(() => {
        setIsEnabled(action ? action.enabled : true)
        setSelectedBatchChange(action && action.__typename === 'MonitorBatchChange' ? action.batchChange.id : '')
    }, [action])

    const onDelete: React.FormEventHandler = useCallback(() => {
        setAction(undefined)
    }, [setAction])

    // TODO: Handle errors.
    const { loading, data } = useQuery<ListBatchChangeAutocompleteResult, ListBatchChangeAutocompleteVariables>(
        LIST_AUTOCOMPLETE,
        {
            variables: {
                namespace: authenticatedUser.id,
            },
        }
    )

    return (
        <ActionEditor
            title={
                <div className="d-flex align-items-center">
                    Re-run a batch change <ProductStatusBadge className="ml-1" status="experimental" />{' '}
                </div>
            }
            label="Re-run batch change"
            subtitle="Triggers a batch change re-execution."
            idName="batchChange"
            disabled={disabled}
            completed={!!action}
            completedSubtitle="The batch change will be re-executed and auto applied."
            actionEnabled={isEnabled}
            toggleActionEnabled={toggleEnabled}
            canSubmit={selectedBatchChange !== undefined}
            onSubmit={onSubmit}
            onCancel={onCancel}
            canDelete={!!action}
            onDelete={onDelete}
            _testStartOpen={_testStartOpen}
        >
            <Alert variant="info" className="mt-4">
                The specified webhook URL will be called with a JSON payload. The format of this JSON payload is still
                being modified. Once it is decided on, documentation will be available.
            </Alert>
            <Select
                id="code-monitor-batch-change-id"
                label="Batch change"
                className="mb-2"
                data-testid="batch-change-id"
                required={true}
                onChange={event => {
                    setSelectedBatchChange(event.target.value)
                }}
                value={selectedBatchChange}
                autoFocus={true}
                spellCheck={false}
                disabled={loading}
            >
                {data?.namespace?.batchChanges.nodes.map(node => (
                    <option key={node.id} value={node.id}>
                        {node.namespace.namespaceName} / {node.name}
                    </option>
                ))}
            </Select>
        </ActionEditor>
    )
}
