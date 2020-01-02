import React, { useState, useEffect } from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { useError } from '../../../../util/useObservable'
import { queryNamespaces } from '../../../namespaces/backend'

interface Props {
    value: GQL.Namespace['id'] | undefined
    onChange: (newValue: GQL.Namespace['id']) => void

    id?: string
    className?: string

    /** For testing only. */
    _queryNamespaces?: typeof queryNamespaces
}

/**
 * A select field for a campaign's namespace.
 */
export const CampaignNamespaceField: React.FunctionComponent<Props> = ({
    value,
    onChange,
    id,
    className = '',
    _queryNamespaces = queryNamespaces,
}) => {
    const [namespaces, setNamespaces] = useState<Pick<GQL.Namespace, 'id' | 'namespaceName'>[]>()

    // For errors during fetching.
    const triggerError = useError()

    useEffect(() => {
        const subscription = _queryNamespaces().subscribe({ next: setNamespaces, error: triggerError })
        return () => subscription.unsubscribe()
    }, [_queryNamespaces, triggerError])

    // Set initial namespace (in case use never interacts with the field and onChange is never called).
    useEffect(() => {
        if (value === undefined && namespaces !== undefined) {
            onChange(namespaces?.[0].id)
        }
    }, [namespaces, onChange, value])

    return (
        <select
            id={id}
            className={`form-control ${className}`}
            required={true}
            value={value || namespaces?.[0].id}
            onChange={event => onChange(event.target.value)}
            disabled={namespaces === undefined}
        >
            {namespaces ? (
                namespaces.map(namespace => (
                    <option value={namespace.id} key={namespace.id}>
                        {namespace.namespaceName}
                    </option>
                ))
            ) : (
                <option disabled={true}>Loading...</option>
            )}
        </select>
    )
}
