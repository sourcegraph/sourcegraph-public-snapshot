import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AddIcon from 'mdi-react/AddIcon'
import React, { useCallback, useState } from 'react'
import { CheckTemplate } from '../../../../../../shared/src/api/client/services/checkTemplates'
import { MultilineTextField } from '../../../../../../shared/src/components/multilineTextField/MultilineTextField'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { Form } from '../../../../components/Form'
import { createThread } from '../../../../discussions/backend'
import { ThreadTitleFormGroup } from '../../../threads/form/ThreadTitleFormGroup'
import { ChecksAreaContext } from '../../global/ChecksArea'

interface Props extends Pick<ChecksAreaContext, 'project'> {
    checkTemplate: CheckTemplate
    className?: string
    history: H.History
}

const LOADING: 'loading' = 'loading'

/**
 * A form to create a new check thread.
 */
export const NewCheckThreadForm: React.FunctionComponent<Props> = ({
    project,
    checkTemplate,
    className = '',
    history,
}) => {
    const [title, setTitle] = useState(checkTemplate.title)
    const onTitleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setTitle(e.currentTarget.value),
        []
    )

    const [body, setBody] = useState('')
    const onBodyChange = useCallback<React.ChangeEventHandler<HTMLTextAreaElement>>(
        e => setBody(e.currentTarget.value),
        []
    )

    const [creationOrError, setCreationOrError] = useState<null | typeof LOADING | GQL.IDiscussionThread | ErrorLike>(
        null
    )
    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setCreationOrError(LOADING)
            try {
                const thread = await createThread({
                    project: project.id,
                    title,
                    contents: body,
                    type: GQL.ThreadType.CHECK,
                    settings: JSON.stringify(checkTemplate.settings || {}, null, 2),
                    status: GQL.ThreadStatus.INACTIVE,
                }).toPromise()
                setCreationOrError(thread)
                history.push(thread.url)
            } catch (err) {
                setCreationOrError(asError(err))
            }
        },
        [project.id, title, body, checkTemplate.settings, history]
    )

    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <ThreadTitleFormGroup value={title} onChange={onTitleChange} disabled={creationOrError === LOADING} />
            <div className="form-group">
                <label
                    className="d-none"
                    htmlFor="new-check-thread-form__body"
                    aria-label="new-check-thread-form__body d-none"
                >
                    Body
                </label>
                <MultilineTextField
                    className="form-control"
                    id="new-check-thread-form__body"
                    minRows={5}
                    value={body}
                    onChange={onBodyChange}
                />
            </div>
            <button type="submit" disabled={creationOrError === LOADING} className="btn btn-primary">
                {creationOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : (
                    <AddIcon className="icon-inline" />
                )}{' '}
                Create check
            </button>
            <small className="form-text text-muted">
                You can preview and customize this check before it performs actions or sends notifications.
            </small>
            {isErrorLike(creationOrError) && <div className="alert alert-danger mt-3">{creationOrError.message}</div>}
        </Form>
    )
}
