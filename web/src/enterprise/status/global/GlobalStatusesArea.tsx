import { StatusScope } from '@sourcegraph/extension-api-classes'
import React from 'react'
import { RouteComponentProps } from 'react-router'
import { StatusesArea, StatusesAreaContext } from '../statusesArea/StatusesArea'

interface Props
    extends Pick<StatusesAreaContext, Exclude<keyof StatusesAreaContext, 'scope'>>,
        RouteComponentProps<{}> {}

/**
 * The global statuses area.
 */
export const GlobalStatusesArea: React.FunctionComponent<Props> = ({ ...props }) => (
    <StatusesArea {...props} scope={StatusScope.Global} />
)
