import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import DotsHorizontalIcon from 'mdi-react/DotsHorizontalIcon'
import React, { useCallback, useEffect, useState } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { ThreadFormTitleField } from './ThreadFormTitleField'

export interface ThreadFormData extends Pick<GQL.IThread, 'title' | 'baseRef' | 'headRef'> {}

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
    initialValue = { title: '', baseRef: 'master' /*TODO!(sqs):un-hardcode*/, headRef: '' },
    onDismiss,
    onSubmit: onSubmitThread,
    buttonText,
    isLoading,
    className = '',
}) => {
    const [title, setTitle] = useState(initialValue.title)
    const onTitleChange = useCallback(value => setTitle(value), [])
    useEffect(() => setTitle(initialValue.title), [initialValue.title])

    const [baseRef, setBaseRef] = useState(initialValue.baseRef)
    const onBaseRefChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setBaseRef(e.currentTarget.value),
        []
    )
    useEffect(() => setBaseRef(initialValue.baseRef), [initialValue.baseRef])

    const [headRef, setHeadRef] = useState(initialValue.headRef)
    const onHeadRefChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setHeadRef(e.currentTarget.value),
        []
    )
    useEffect(() => setHeadRef(initialValue.headRef), [initialValue.headRef])

    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            onSubmitThread({ title, baseRef, headRef })
        },
        [onSubmitThread, title, baseRef, headRef]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <div className="form-row align-items-end">
                <ThreadFormTitleField value={title} onChange={onTitleChange} autoFocus={true} />
                <div className="form-group">
                    <label htmlFor="thread-form__baseRef">Range</label>
                    <div className="input-group align-items-center">
                        <input
                            type="text"
                            id="thread-form__baseRef"
                            className="form-control"
                            required={true}
                            placeholder="Base ref"
                            value={baseRef}
                            onChange={onBaseRefChange}
                        />
                        <DotsHorizontalIcon className="icon-inline mx-2" />
                        <input
                            type="text"
                            id="thread-form__headRef"
                            className="form-control"
                            required={true}
                            placeholder="Head ref"
                            value={headRef}
                            onChange={onHeadRefChange}
                        />
                    </div>
                </div>
                <div className="form-group text-right">
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
