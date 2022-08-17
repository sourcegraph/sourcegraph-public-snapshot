import React, { useCallback } from 'react'

import { mdiInformationOutline } from '@mdi/js'

import { SettingsOrgSubject, SettingsUserSubject } from '@sourcegraph/shared/src/settings/settings'
import { Icon, Select, Tooltip } from '@sourcegraph/wildcard'

type PartialNamespace =
    | Pick<SettingsUserSubject, '__typename' | 'id' | 'username' | 'displayName'>
    | Pick<SettingsOrgSubject, '__typename' | 'id' | 'name' | 'displayName'>

const getNamespaceDisplayName = (namespace: PartialNamespace): string => {
    switch (namespace.__typename) {
        case 'User':
            return namespace.displayName ?? namespace.username
        case 'Org':
            return namespace.displayName ?? namespace.name
    }
}

const NAMESPACE_SELECTOR_ID = 'batch-spec-execution-namespace-selector'

type NamespaceSelectorProps = {
    namespaces: PartialNamespace[]
    selectedNamespace: string
} & ({ disabled: true; onSelect?: undefined } | { disabled?: false; onSelect: (namespace: PartialNamespace) => void }) // Either the selector is disabled and there's on onSelect, or the selector is enabled and there is one.

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
            label={
                <>
                    <strong className="text-nowrap mb-2">Namespace</strong>
                    <Tooltip content="Coming soon">
                        <Icon aria-label="Coming soon" className="ml-1" svgPath={mdiInformationOutline} />
                    </Tooltip>
                </>
            }
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
