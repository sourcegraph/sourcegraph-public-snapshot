import React, { useState, useCallback } from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Form } from '../../../../components/Form'
import { importThreadsFromExternalService } from '../../../campaigns/detail/threads/ImportThreadsFromExternalServiceToCampaignDropdownButton'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import { NotificationType } from '@sourcegraph/extension-api-classes'

interface Props extends ExtensionsControllerNotificationProps {
    onThreadsSelect: (threads: Pick<GQL.IThread, 'id'>[]) => void

    className?: string
    disabled?: boolean
}

/**
 * A form to import threads by query from an external service.
 */
export const ImportThreadsByQueryFromExternalServiceForm: React.FunctionComponent<Props> = ({
    onThreadsSelect,
    className = '',
    disabled,
    extensionsController,
}) => {
    const [query, setQuery] = useState('')
    const onQueryChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setQuery(e.currentTarget.value)
    }, [])

    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                const threads = await importThreadsFromExternalService({ byQuery: query })
                setIsLoading(false)
                onThreadsSelect(threads)
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error importing threads by query from external service: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, onThreadsSelect, query]
    )

    disabled = disabled || isLoading

    return (
        <Form onSubmit={onSubmit} className={className}>
            <div className="form-group">
                <label htmlFor="import-threads-by-query-from-external-service-form__query">
                    Issue/pull request query
                </label>
                <input
                    type="text"
                    id="import-threads-by-query-from-external-service-form__query"
                    className="form-control"
                    required={true}
                    minLength={1}
                    placeholder="org:acme label:a"
                    value={query}
                    onChange={onQueryChange}
                    autoFocus={true}
                    disabled={disabled}
                />
            </div>
            <button type="submit" className="btn btn-primary" disabled={disabled}>
                Add
            </button>
        </Form>
    )
}
