import React, { useCallback } from 'react'

import { Button, Tooltip } from '@sourcegraph/wildcard'

import type { ExecutorSecretScope } from '../../../graphql-operations'

export interface Props {
    scope: ExecutorSecretScope
    label: string
    description: string
    selected: boolean
    onSelect: (scope: ExecutorSecretScope) => void

    className?: string
}

export const ExecutorSecretScopeSelector: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    scope,
    className,
    selected,
    onSelect,
    label,
    description,
}) => {
    const onClick = useCallback(() => {
        onSelect(scope)
    }, [onSelect, scope])

    return (
        <Tooltip content={description}>
            <Button className={className} onClick={onClick} disabled={selected} outline={true} variant="secondary">
                {label}
            </Button>
        </Tooltip>
    )
}
