import React, { useCallback, useState } from 'react'
import { NotificationType } from '../../../../../../shared/src/api/client/services/notifications'
import { ExtensionsControllerNotificationProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { Form } from '../../../../components/Form'
import { updateChangeset } from '../../common/EditChangesetForm'

interface Props extends ExtensionsControllerNotificationProps {
    changeset: Pick<GQL.IChangeset, 'id' | 'number' | 'title'>
    onChangesetUpdate: () => void
    className?: string

    /** Provided by tests only. */
    _updateChangeset?: typeof updateChangeset
}

/**
 * The title in the changeset header, which has an edit mode.
 */
// tslint:disable: jsx-no-lambda
export const ChangesetHeaderEditableTitle: React.FunctionComponent<Props> = ({
    changeset,
    onChangesetUpdate,
    className = '',
    extensionsController,
    _updateChangeset = updateChangeset,
}) => {
    const [state, setState] = useState<'viewing' | 'editing' | 'loading'>('viewing')
    const [uncommittedTitle, setUncommittedTitle] = useState(changeset.title)

    const onSubmit: React.FormEventHandler = async e => {
        e.preventDefault()
        setState('loading')
        try {
            await _updateChangeset({ id: changeset.id, title: uncommittedTitle })
            onChangesetUpdate()
        } catch (err) {
            extensionsController.services.notifications.showMessages.next({
                message: `Error editing changeset title: ${err.message}`,
                type: NotificationType.Error,
            })
        } finally {
            setState('viewing')
        }
    }

    const onEditClick = useCallback(() => setState('editing'), [])
    const onCancelClick = useCallback<React.MouseEventHandler>(
        e => {
            e.preventDefault()
            setState('viewing')
            setUncommittedTitle(changeset.title)
        },
        [changeset.title]
    )

    return state === 'viewing' ? (
        <div className={`d-flex align-items-start justify-content-between ${className}`}>
            <h1 className="font-weight-normal mb-0 h2 mr-2">
                {changeset.title} <span className="text-muted font-weight-normal">#{changeset.number}</span>
            </h1>
            <button type="button" className="btn btn-secondary btn-sm mt-1" onClick={onEditClick}>
                Edit
            </button>
        </div>
    ) : (
        <Form className={`form d-flex ${className}`} onSubmit={onSubmit}>
            <input
                type="text"
                className="form-control flex-1 mr-2"
                value={uncommittedTitle}
                onChange={e => setUncommittedTitle(e.currentTarget.value)}
                placeholder="Title"
                autoComplete="off"
                autoFocus={true}
                disabled={state === 'loading'}
            />
            <div className="text-nowrap flex-0 d-flex align-items-center">
                <button type="submit" className="btn btn-success" disabled={state === 'loading'}>
                    Save
                </button>
                <button type="reset" className="btn btn-link" onClick={onCancelClick} disabled={state === 'loading'}>
                    Cancel
                </button>
            </div>
        </Form>
    )
}
