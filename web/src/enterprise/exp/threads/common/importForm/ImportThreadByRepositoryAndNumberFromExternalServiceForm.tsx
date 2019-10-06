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
 * A form to import a thread by repository name and number from an external service.
 */
export const ImportThreadByRepositoryAndNumberFromExternalServiceForm: React.FunctionComponent<Props> = ({
    onThreadsSelect,
    className = '',
    disabled,
    extensionsController,
}) => {
    const [repositoryName, setRepositoryName] = useState('')
    const onRepositoryNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setRepositoryName(e.currentTarget.value)
    }, [])

    const [number, setNumber] = useState('')
    const onNumberChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(e => {
        setNumber(e.currentTarget.value)
    }, [])

    const [isLoading, setIsLoading] = useState(false)
    const onSubmit = useCallback<React.FormEventHandler<HTMLFormElement>>(
        async e => {
            e.preventDefault()
            setIsLoading(true)
            try {
                const threads = await importThreadsFromExternalService({
                    byRepositoryAndNumber: { repositoryName, number: parseInt(number, 10) },
                })
                setIsLoading(false)
                onThreadsSelect(threads)
            } catch (err) {
                setIsLoading(false)
                extensionsController.services.notifications.showMessages.next({
                    message: `Error importing thread from external service: ${err.message}`,
                    type: NotificationType.Error,
                })
            }
        },
        [extensionsController.services.notifications.showMessages, number, onThreadsSelect, repositoryName]
    )

    disabled = disabled || isLoading

    return (
        <Form onSubmit={onSubmit} className={className}>
            <div className="form-group">
                <label htmlFor="import-thread-by-repository-and-number-from-external-service-form__repositoryName">
                    Repository
                </label>
                <input
                    type="text"
                    id="import-thread-by-repository-and-number-from-external-service-form__repositoryName"
                    className="form-control"
                    required={true}
                    minLength={1}
                    placeholder="myorg/myrepo"
                    value={repositoryName}
                    onChange={onRepositoryNameChange}
                    autoFocus={true}
                    disabled={disabled}
                />
            </div>
            <div className="form-group">
                <label htmlFor="import-thread-by-repository-and-number-from-external-service-form__number">
                    Number
                </label>
                <input
                    type="number"
                    id="import-thread-by-repository-and-number-from-external-service-form__number"
                    className="form-control"
                    required={true}
                    step="any"
                    placeholder="123"
                    value={number}
                    onChange={onNumberChange}
                    disabled={disabled}
                />
            </div>
            <button type="submit" className="btn btn-primary" disabled={disabled}>
                Add
            </button>
        </Form>
    )
}
