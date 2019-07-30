import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useEffect, useState } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'

export interface ChangesetFormData extends Pick<GQL.IChangeset, 'title'> {}

interface Props {
    initialValue?: ChangesetFormData

    /** Called when the form is dismissed with no action taken. */
    onDismiss?: () => void

    /** Called when the form is submitted. */
    onSubmit: (changeset: ChangesetFormData) => void

    buttonText: string
    isLoading: boolean

    className?: string
}

/**
 * A form to create or edit a changeset.
 */
export const ChangesetForm: React.FunctionComponent<Props> = ({
    initialValue = { title: '' },
    onDismiss,
    onSubmit: onSubmitChangeset,
    buttonText,
    isLoading,
    className = '',
}) => {
    const [title, setTitle] = useState(initialValue.title)
    const onNameChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setTitle(e.currentTarget.value),
        []
    )
    useEffect(() => setTitle(initialValue.title), [initialValue.title])

    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            onSubmitChangeset({ title })
        },
        [title, onSubmitChangeset]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <div className="form-row align-items-end">
                <div className="form-group mb-md-0 col-md-3">
                    <label htmlFor="changeset-form__name">Name</label>
                    <input
                        type="text"
                        id="changeset-form__name"
                        className="form-control"
                        required={true}
                        minLength={1}
                        placeholder="Changeset title"
                        value={title}
                        onChange={onNameChange}
                        autoFocus={true}
                    />
                </div>
                <div className="form-group mb-md-0 col-md-3 text-right">
                    {onDismiss && (
                        <button type="reset" className="btn btn-secondary mr-2" onClick={onDismiss}>
                            Cancel
                        </button>
                    )}
                    <button type="submit" disabled={isLoading} className="btn btn-success">
                        {isLoading ? <LoadingSpinner className="icon-inline" /> : buttonText}
                    </button>
                </div>
            </div>
        </Form>
    )
}
