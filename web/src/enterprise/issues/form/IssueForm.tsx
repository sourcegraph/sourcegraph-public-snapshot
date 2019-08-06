import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useEffect, useState } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { Form } from '../../../components/Form'
import { ThreadlikeFormTitleField } from '../../threadlike/form/ThreadlikeFormTitleField'

export interface IssueFormData extends Pick<GQL.IIssue, 'title'> {}

interface Props {
    initialValue?: IssueFormData

    /** Called when the form is dismissed with no action taken. */
    onDismiss?: () => void

    /** Called when the form is submitted. */
    onSubmit: (issue: IssueFormData) => void

    buttonText: string
    isLoading: boolean

    className?: string
}

/**
 * A form to create or edit a issue.
 */
export const IssueForm: React.FunctionComponent<Props> = ({
    initialValue = { title: '' },
    onDismiss,
    onSubmit: onSubmitIssue,
    buttonText,
    isLoading,
    className = '',
}) => {
    const [title, setTitle] = useState(initialValue.title)
    const onTitleChange = useCallback(value => setTitle(value), [])
    useEffect(() => setTitle(initialValue.title), [initialValue.title])

    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            onSubmitIssue({ title })
        },
        [onSubmitIssue, title]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <div className="form-row align-items-end">
                <ThreadlikeFormTitleField value={title} onChange={onTitleChange} autoFocus={true} />
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
