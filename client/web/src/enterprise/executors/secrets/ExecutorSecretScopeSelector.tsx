import React, { useCallback } from 'react'

import classNames from 'classnames'

import { Button } from '@sourcegraph/wildcard'

import { ExecutorSecretScope } from '../../../graphql-operations'

import styles from './ExecutorSecretScopeSelector.module.scss'

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
        <Button
            className={classNames(styles.selector, 'px-4', 'py-3', selected && styles.selected, className)}
            onClick={onClick}
            outline={true}
            variant="secondary"
        >
            <Text className={classNames(styles.count)}>{label}</Text>
            <span className={classNames('text-muted')}>{description}</span>
        </Button>
    )
}
