import H from 'history'
import React from 'react'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import { Comment } from '../../comments/Comment'
import { ChangesetAreaContext } from './ChangesetArea'
import { ChangesetHeaderEditableTitle } from './header/ChangesetHeaderEditableTitle'

interface Props extends Pick<ChangesetAreaContext, 'changeset' | 'onChangesetUpdate'>, ExtensionsControllerProps {
    className?: string

    history: H.History
}

/**
 * The overview for a single changeset.
 */
export const ChangesetOverview: React.FunctionComponent<Props> = ({
    changeset,
    onChangesetUpdate,
    className = '',
    ...props
}) => (
    <div className={`changeset-overview ${className || ''}`}>
        <ChangesetHeaderEditableTitle
            {...props}
            changeset={changeset}
            onChangesetUpdate={onChangesetUpdate}
            className="mb-3"
        />
        <Comment
            {...props}
            comment={changeset}
            onCommentUpdate={onChangesetUpdate}
            createdVerb="opened changeset"
            emptyBody="No description provided."
            className="mb-3"
        />
    </div>
)
