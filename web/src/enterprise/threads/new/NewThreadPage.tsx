import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AddIcon from 'mdi-react/AddIcon'
import React, { useCallback, useState } from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { Form } from '../../../components/Form'
import { ModalPage } from '../../../components/ModalPage'
import { PageTitle } from '../../../components/PageTitle'
import { createThread } from '../../../discussions/backend'
import { ThreadTitleFormGroup } from '../form/ThreadTitleFormGroup'
import { ThreadsAreaContext } from '../global/ThreadsArea'
import { threadNoun } from '../util'

interface Props extends ThreadsAreaContext {
    history: H.History
}

const LOADING: 'loading' = 'loading'

/**
 * Shows a form to create a new thread.
 */
export const NewThreadPage: React.FunctionComponent<Props> = ({ type, history }) => {
    const [title, setTitle] = useState('')
    const onTitleChange = useCallback<React.ChangeEventHandler<HTMLInputElement>>(
        e => setTitle(e.currentTarget.value),
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
                const thread = await createThread({ title, contents: '', type: GQL.ThreadType.THREAD }).toPromise()
                setCreationOrError(thread)
                history.push(thread.url)
            } catch (err) {
                setCreationOrError(asError(err))
            }
        },
        [title, history]
    )

    return (
        <>
            <PageTitle title="New thread" />
            <ModalPage>
                <h2>New {threadNoun(type)}</h2>
                <Form onSubmit={onSubmit}>
                    <ThreadTitleFormGroup
                        value={title}
                        onChange={onTitleChange}
                        disabled={creationOrError === LOADING}
                    />
                    <button type="submit" disabled={creationOrError === LOADING} className="btn btn-primary">
                        {creationOrError === LOADING ? (
                            <LoadingSpinner className="icon-inline" />
                        ) : (
                            <AddIcon className="icon-inline" />
                        )}{' '}
                        Create {threadNoun(type)}
                    </button>
                </Form>
                {isErrorLike(creationOrError) && (
                    <div className="alert alert-danger mt-3">{creationOrError.message}</div>
                )}
            </ModalPage>
        </>
    )
}
