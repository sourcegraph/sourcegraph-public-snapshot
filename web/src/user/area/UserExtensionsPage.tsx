import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { ConfiguredExtensionsPage } from '../../registry/ConfiguredExtensionsPage'
import { UserAreaPageProps } from './UserArea'

interface Props extends UserAreaPageProps, RouteComponentProps<{}> {}

/** Displays extensions used by this user. */
export const UserExtensionsPage: React.SFC<Props> = props => (
    <ConfiguredExtensionsPage {...props} publisher={props.user} subject={props.user} showUserActions={true} />
)
