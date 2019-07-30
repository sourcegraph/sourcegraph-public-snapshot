import ExternalLinkIcon from 'mdi-react/ExternalLinkIcon'
import React from 'react'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { ChangesetAreaContext } from './ChangesetArea'
import { ChangesetHeaderEditableTitle } from './header/ChangesetHeaderEditableTitle'

interface Props
    extends Pick<ChangesetAreaContext, 'changeset' | 'onChangesetUpdate'>,
        ExtensionsControllerNotificationProps {
    className?: string
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
        {changeset.externalURL && (
            <a href={changeset.externalURL}>
                <ExternalLinkIcon className="icon-inline mr-1" /> View pull request
            </a>
        )}
    </div>
)
