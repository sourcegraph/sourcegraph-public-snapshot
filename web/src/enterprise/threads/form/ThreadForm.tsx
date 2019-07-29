import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useEffect, useState } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'

export interface ThreadFormData extends Pick<GQL.IThread, 'title'> {}

interface Props {
    initialValue?: ThreadFormData

    /** Called when the form is dismissed with no action taken. */
    onDismiss?: () => void

    /** Called when the form is submitted. */
    onSubmit: (thread: ThreadFormData) => void

    buttonText: string
    isLoading: boolean

    className?: string
}

/**
 * A form to create or edit a thread.
 */
export const ThreadForm: React.FunctionComponent<Props> = ({
    initialValue = { title: '' },
    onDismiss,
    onSubmit: onSubmitThread,
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
            onSubmitThread({ title })
        },
        [title, onSubmitThread]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <div className="form-row align-items-end">
                <div className="form-group mb-md-0 col-md-3">
                    <label htmlFor="thread-form__name">Name</label>
                    <input
                        type="text"
                        id="thread-form__name"
                        className="form-control"
                        required={true}
                        minLength={1}
                        placeholder="Thread title"
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
