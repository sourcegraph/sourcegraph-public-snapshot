import React, { useCallback } from 'react'

import { SettingsOrgSubject, SettingsUserSubject } from '@sourcegraph/shared/src/settings/settings'
import { Select } from '@sourcegraph/wildcard'

type PartialNamespace =
    | Pick<SettingsUserSubject, '__typename' | 'id' | 'username' | 'displayName'>
    | Pick<SettingsOrgSubject, '__typename' | 'id' | 'name' | 'displayName'>

const getNamespaceDisplayName = (namespace: PartialNamespace): string => {
    switch (namespace.__typename) {
        case 'User':
            return namespace.displayName ? namespace.displayName : namespace.username
        case 'Org':
            return namespace.displayName ? namespace.displayName : namespace.name
    }
}

const NAMESPACE_SELECTOR_ID = 'batch-spec-execution-namespace-selector'

interface NamespaceSelectorProps {
    namespaces: PartialNamespace[]
    selectedNamespace: string
    disabled?: boolean
    onSelect: (namespace: PartialNamespace) => void
}

export const NamespaceSelector: React.FunctionComponent<React.PropsWithChildren<NamespaceSelectorProps>> = ({
    namespaces,
    disabled,
    selectedNamespace,
    onSelect,
}) => {
    const onSelectNamespace = useCallback<React.ChangeEventHandler<HTMLSelectElement>>(
        event => {
            if (disabled) {
                return
            }

            const selectedNamespace = namespaces.find(
                (namespace): namespace is SettingsUserSubject | SettingsOrgSubject =>
                    namespace.id === event.target.value
            )
            onSelect(selectedNamespace || namespaces[0])
        },
        [disabled, onSelect, namespaces]
    )

    return (
        <Select
            label={<strong className="text-nowrap mb-2">Namespace</strong>}
            isCustomStyle={true}
            id={NAMESPACE_SELECTOR_ID}
            value={selectedNamespace}
            onChange={onSelectNamespace}
            disabled={disabled}
        >
            {namespaces.map(namespace => (
                <option key={namespace.id} value={namespace.id}>
                    {getNamespaceDisplayName(namespace)}
                </option>
            ))}
        </Select>
    )
}
