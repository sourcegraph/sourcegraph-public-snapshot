import classNames from 'classnames'
import CheckBoldIcon from 'mdi-react/CheckBoldIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'

import { ComponentStateFields, ComponentStatusState } from '../../../../../../graphql-operations'
import { STATE_TO_COLOR } from '../../../component/OverviewStatusContextItem'

export const ComponentStateIndicator: React.FunctionComponent<{
    entity: ComponentStateFields
    className?: string
}> = ({ entity, className }) => (
    <span className={classNames(`text-${STATE_TO_COLOR[entity.status.state]}`, className)}>
        {entity.status.state === ComponentStatusState.SUCCESS ? (
            <CheckBoldIcon className="icon-inline" />
        ) : entity.status.state === ComponentStatusState.FAILURE ||
          entity.status.state === ComponentStatusState.ERROR ? (
            <CloseIcon className="icon-inline" />
        ) : (
            entity.status.state.toLowerCase()
        )}
    </span>
)
