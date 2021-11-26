import classNames from 'classnames'
import CheckBoldIcon from 'mdi-react/CheckBoldIcon'
import CloseIcon from 'mdi-react/CloseIcon'
import React from 'react'
import { useQuery, gql } from '@sourcegraph/http-client'

import { ComponentStateIndicatorFields, ComponentStatusState } from '../../../../../../graphql-operations'
import { STATE_TO_COLOR } from '../../../component/OverviewStatusContextItem'

export const COMPONENT_STATE_INDICATOR_FRAGMENT = gql`
    fragment ComponentStateIndicatorFields on Component {
        status {
            state
        }
    }
`

export const ComponentStateIndicator: React.FunctionComponent<{
    component: ComponentStateIndicatorFields
    className?: string
}> = ({ component, className }) => (
    <span className={classNames(`text-${STATE_TO_COLOR[component.status.state]}`, className)}>
        {component.status.state === ComponentStatusState.SUCCESS ? (
            <CheckBoldIcon className="icon-inline" />
        ) : component.status.state === ComponentStatusState.FAILURE ||
          component.status.state === ComponentStatusState.ERROR ? (
            <CloseIcon className="icon-inline" />
        ) : (
            component.status.state.toLowerCase()
        )}
    </span>
)
