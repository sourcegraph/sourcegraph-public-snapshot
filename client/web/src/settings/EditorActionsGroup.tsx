import React, { useEffect } from 'react'

import classNames from 'classnames'
import type * as jsonc from 'jsonc-parser'
import { useSearchParams } from 'react-router-dom'

import { Button, Text } from '@sourcegraph/wildcard'

import styles from './EditorActionsGroup.module.scss'

/**
 * A helper function that modifies site configuration to configure specific
 * common things, such as syncing GitHub repositories.
 */
export type ConfigInsertionFunction = (configJSON: string) => {
    /** The edits to make to the input configuration to insert the new configuration. */
    edits: jsonc.Edit[]

    /** Select text in inserted JSON. */
    selectText?: string | number

    /**
     * If set, the selection is an empty selection that begins at the left-hand match of selectText plus this
     * offset. For example, if selectText is "foo" and cursorOffset is 2, then the final selection will be a cursor
     * "|" positioned as "fo|o".
     */
    cursorOffset?: number
}

export interface EditorAction {
    id: string
    label: string
    run: ConfigInsertionFunction
}

export interface EditorActionsGroupProps {
    actions: EditorAction[]
    onClick: (id: string) => void
    actionsAvailable: boolean
}

export const EditorActionsGroup: React.FunctionComponent<EditorActionsGroupProps> = ({
    actions,
    onClick,
    actionsAvailable,
}) => {
    const [queryParameters, setSearchParams] = useSearchParams()
    const id = queryParameters.get('actionItem')

    useEffect(() => {
        if (id && actionsAvailable) {
            queryParameters.delete('actionItem')
            setSearchParams(queryParameters.toString())
            onClick(id)
        }
    }, [id, setSearchParams, queryParameters, actionsAvailable, onClick])

    return (
        <>
            {actions.length > 0 && (
                <Text className="mb-1">
                    <strong>Quick actions:</strong>
                </Text>
            )}
            <div className={classNames(styles.actions, 'mb-2')}>
                {actions.map(({ id, label }) => (
                    <Button key={id} onClick={() => onClick(id)} variant="secondary" outline={true} size="sm">
                        {label}
                    </Button>
                ))}
            </div>
        </>
    )
}
