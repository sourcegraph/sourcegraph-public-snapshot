import H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { CheckTemplate } from '../../../../../../shared/src/api/client/services/checkTemplates'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import { CheckTemplateItem } from '../../components/CheckTemplateItem'

interface Props extends ExtensionsControllerProps<'services'> {
    urlForCheckTemplate: (checkTemplateId: string) => H.LocationDescriptor
}

/**
 * A list of available templates for a new check.
 */
export const CheckThreadTemplateSelectFormControl: React.FunctionComponent<Props> = ({
    urlForCheckTemplate,
    extensionsController,
}) => {
    const [query, setQuery] = useState('')
    const onQueryChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setQuery(e.currentTarget.value),
        []
    )

    const [checkTemplates, setCheckTemplates] = useState<CheckTemplate[]>()
    useEffect(() => {
        const subscription = extensionsController.services.checkTemplates
            .getCheckTemplates()
            .subscribe(checkTemplates => setCheckTemplates(checkTemplates || undefined))
        return () => subscription.unsubscribe()
    }, [extensionsController.services.checkTemplates])

    return checkTemplates ? (
        checkTemplates.length === 0 ? (
            <small>No check templates found.</small>
        ) : (
            <ul className="list-group">
                <li className="list-group-item p-0">
                    <input
                        type="search"
                        className="form-control border-0 px-3 py-2 rounded-bottom-0"
                        value={query}
                        onChange={onQueryChange}
                        placeholder="Search"
                    />
                </li>
                {checkTemplates
                    .filter(
                        ({ title, description }) =>
                            title.toLowerCase().includes(query.toLowerCase()) ||
                            (description && description.toLowerCase().includes(query.toLowerCase()))
                    )
                    .map((t, i) => (
                        <CheckTemplateItem
                            key={i}
                            element="li"
                            checkTemplate={t}
                            className="list-group-item list-group-item-action position-relative"
                            endFragment={
                                <Link
                                    to={urlForCheckTemplate(t.id)}
                                    className="stretched-link mb-0 text-decoration-none d-flex align-items-center text-body"
                                />
                            }
                        />
                    ))}
            </ul>
        )
    ) : null
}
