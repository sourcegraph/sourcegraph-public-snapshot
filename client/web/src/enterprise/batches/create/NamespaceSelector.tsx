import React, { useCallback } from 'react'

import type { OrgSettingFields, UserSettingFields } from '@sourcegraph/shared/src/graphql-operations'
import { Select } from '@sourcegraph/wildcard'

type PartialNamespace =
    | Pick<UserSettingFields, '__typename' | 'id' | 'username' | 'displayName'>
    | Pick<OrgSettingFields, '__typename' | 'id' | 'name' | 'displayName'>

const getNamespaceDisplayName = (namespace: PartialNamespace): string => {
    switch (namespace.__typename) {
        case 'User': {
            return namespace.displayName ? namespace.displayName : namespace.username
        }
        case 'Org': {
            return namespace.displayName ? namespace.displayName : namespace.name
        }
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
                (namespace): namespace is UserSettingFields | OrgSettingFields => namespace.id === event.target.value
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
