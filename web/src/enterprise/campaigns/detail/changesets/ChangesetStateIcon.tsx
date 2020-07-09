import { changesetStateIcons, changesetStatusColorClasses, changesetStateLabels } from './presentation'
import { ChangesetState } from '../../../../../../shared/src/graphql/schema'
import React from 'react'
import classNames from 'classnames'

export interface ChangesetStateIconProps {
    state: ChangesetState
}

export const ChangesetStateIcon: React.FunctionComponent<ChangesetStateIconProps> = ({ state }) => {
    const Icon = changesetStateIcons[state]
    return (
        <Icon
            className={classNames('mr-1 icon-inline', `text-${changesetStatusColorClasses[state]}`)}
            data-tooltip={changesetStateLabels[state]}
        />
    )
}
