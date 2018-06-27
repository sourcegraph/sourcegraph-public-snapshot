import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { ExtensionsProps } from '../../backend/features'
import { ConfiguredExtensionsPage } from '../../registry/ConfiguredExtensionsPage'
import { UserAreaPageProps } from './UserArea'

interface Props extends UserAreaPageProps, RouteComponentProps<{}>, ExtensionsProps {}

/** Displays extensions used by this user. */
export const UserExtensionsPage: React.SFC<Props> = props => (
    <ConfiguredExtensionsPage
        {...props}
        publisher={props.user}
        subject={props.user}
        showUserActions={true}
        updateOnChange={props.extensions}
    />
)
