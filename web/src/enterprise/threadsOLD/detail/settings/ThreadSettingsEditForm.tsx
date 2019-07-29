import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import CheckIcon from 'mdi-react/CheckIcon'
import React, { useCallback, useState } from 'react'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike, isErrorLike } from '../../../../../../shared/src/util/errors'
import { Form } from '../../../../components/Form'
import { updateThread } from '../../../../discussions/backend'
import { ThreadSettingsEditor } from '../../form/ThreadSettingsEditor'

interface Props {
    thread: Pick<GQL.IDiscussionThread, 'id' | 'settings'>
    onThreadUpdate: (thread: GQL.IDiscussionThread) => void
    className?: string
    isLightTheme: boolean
    history: H.History
}

const LOADING: 'loading' = 'loading'

export const ThreadSettingsEditForm: React.FunctionComponent<Props> = ({
    thread: { id: threadID, settings },
    onThreadUpdate,
    className = '',
    ...props
}) => {
    const [uncommittedSettings, setUncommittedSettings] = useState(settings)
    const [updateOrError, setUpdateOrError] = useState<null | typeof LOADING | GQL.IDiscussionThread | ErrorLike>(null)
    const onSubmit = useCallback<React.FormEventHandler>(
        async e => {
            e.preventDefault()
            setUpdateOrError(LOADING)
            try {
                const thread = await updateThread({ threadID, settings: uncommittedSettings })
                setUpdateOrError(thread)
                onThreadUpdate(thread)
            } catch (err) {
                setUpdateOrError(asError(err))
            }
        },
        [onThreadUpdate, threadID, uncommittedSettings]
    )
    return (
        <Form className={`form ${className}`} onSubmit={onSubmit}>
            <ThreadSettingsEditor
                {...props}
                value={uncommittedSettings}
                onChange={setUncommittedSettings}
                loading={updateOrError === LOADING}
            />
            <button type="submit" disabled={updateOrError === LOADING} className="btn btn-success">
                {updateOrError === LOADING ? (
                    <LoadingSpinner className="icon-inline" />
                ) : (
                    <CheckIcon className="icon-inline" />
                )}{' '}
                Save
            </button>
            {isErrorLike(updateOrError) && <div className="alert alert-danger mt-3">{updateOrError.message}</div>}
        </Form>
    )
}
