import classNames from 'classnames'
import React, { useCallback } from 'react'

import { SettingsOrgSubject, SettingsUserSubject } from '@sourcegraph/shared/src/settings/settings'

import styles from './NamespaceSelector.module.scss'

const getNamespaceDisplayName = (namespace: SettingsUserSubject | SettingsOrgSubject): string => {
    switch (namespace.__typename) {
        case 'User':
            return namespace.displayName ?? namespace.username
        case 'Org':
            return namespace.displayName ?? namespace.name
    }
}

const NAMESPACE_SELECTOR_ID = 'batch-spec-execution-namespace-selector'

interface NamespaceSelectorProps {
    namespaces: (SettingsUserSubject | SettingsOrgSubject)[]
    selectedNamespace: string
    onSelect: (namespace: SettingsUserSubject | SettingsOrgSubject) => void
}

export const NamespaceSelector: React.FunctionComponent<NamespaceSelectorProps> = ({
    namespaces,
    selectedNamespace,
    onSelect,
}) => {
    const onSelectNamespace = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => {
            const selectedNamespace = namespaces.find(
                (namespace): namespace is SettingsUserSubject | SettingsOrgSubject =>
                    namespace.id === event.target.value
            )
            onSelect(selectedNamespace || namespaces[0])
        },
        [onSelect, namespaces]
    )

    return (
        <div className="form-group">
            <label className="text-nowrap mb-2" htmlFor={NAMESPACE_SELECTOR_ID}>
                <strong>Namespace:</strong>
            </label>
            <select
                className={classNames(styles.namespaceSelector, 'form-control')}
                id={NAMESPACE_SELECTOR_ID}
                value={selectedNamespace}
                onChange={onSelectNamespace}
            >
                {namespaces.map(namespace => (
                    <option key={namespace.id} value={namespace.id}>
                        {getNamespaceDisplayName(namespace)}
                    </option>
                ))}
            </select>
        </div>
    )
}
