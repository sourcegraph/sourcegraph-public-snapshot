import H from 'history'
import React, { useCallback, useEffect, useState } from 'react'
import TextareaAutosize from 'react-textarea-autosize'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { Form } from '../../components/Form'

interface Props extends ExtensionsControllerProps {
    /** The initial body (used when editing an existing comment). */
    initialBody?: string

    placeholder: string

    /** The label to display on the submit button. */
    submitLabel: string

    /** Called when the submit button is clicked. */
    onSubmit: (body: string) => void

    /**
     * If set, a "Cancel" button is shown, and this callback is called when it is clicked.
     */
    onCancel?: () => void

    disabled?: boolean
    className?: string
    // TODO!(sqs): confirm navigation away when field is dirty
    history: H.History
}

// TODO!(sqs): make this support text field completion in extension api
export const CommentForm: React.FunctionComponent<Props> = ({
    initialBody,
    submitLabel,
    placeholder,
    onSubmit,
    onCancel,
    disabled,
    className = '',
    history,
}) => {
    const [uncommittedBody, setUncommittedBody] = useState(initialBody || '')
    const onChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => setUncommittedBody(e.currentTarget.value),
        []
    )

    const onFormSubmit = useCallback<React.FormEventHandler>(
        e => {
            e.preventDefault()
            onSubmit(uncommittedBody)
        },
        [onSubmit, uncommittedBody]
    )

    useEffect(() => {
        const isDirty = uncommittedBody !== initialBody
        if (isDirty) {
            return history.block('Discard unsaved comment?')
        }
        return undefined
    }, [history, initialBody, uncommittedBody])

    return (
        <Form className={`comment-form ${className}`} onSubmit={onFormSubmit}>
            <TextareaAutosize
                className="form-control"
                placeholder={placeholder}
                value={uncommittedBody}
                onChange={onChange}
                minRows={5}
                autoFocus={true}
                disabled={disabled}
            />
            <div className="d-flex align-items-center justify-content-end">
                {onCancel && (
                    <button type="reset" className="btn btn-link" disabled={disabled} onClick={onCancel}>
                        Cancel
                    </button>
                )}
                <button type="submit" className="btn btn-primary" disabled={disabled}>
                    {submitLabel}
                </button>
            </div>
        </Form>
    )
}
