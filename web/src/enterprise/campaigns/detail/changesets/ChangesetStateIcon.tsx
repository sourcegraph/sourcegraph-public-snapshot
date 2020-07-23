import { changesetExternalStateIcons, changesetExternalStateColorClasses, changesetStateLabels } from './presentation'
import { ChangesetExternalState } from '../../../../../../shared/src/graphql/schema'
import React from 'react'
import classNames from 'classnames'

export interface ChangesetStateIconProps {
    externalState: ChangesetExternalState
}

export const ChangesetStateIcon: React.FunctionComponent<ChangesetStateIconProps> = ({ externalState }) => {
    const Icon = changesetExternalStateIcons[externalState]
    return (
        <Icon
            className={classNames('mr-1 icon-inline', `text-${changesetExternalStateColorClasses[externalState]}`)}
            data-tooltip={changesetStateLabels[externalState]}
        />
    )
}
