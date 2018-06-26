import * as React from 'react'
import { RouteComponentProps } from 'react-router'
import { ConfiguredExtensionsPage } from '../../registry/ConfiguredExtensionsPage'
import { OrgAreaPageProps } from './OrgArea'

interface Props extends OrgAreaPageProps, RouteComponentProps<{}> {}

/** Displays extensions used by this org. */
export const OrgExtensionsPage: React.SFC<Props> = props => (
    <ConfiguredExtensionsPage {...props} publisher={props.org} subject={props.org} showUserActions={true} />
)
