import React from 'react'
import { UserAddCodeHostsPageProps, UserAddCodeHostsPage } from './codeHosts/UserAddCodeHostsPage'
import { codeHostExternalServices } from '../../components/externalServices/externalServices'

export interface UserAddCodeHostsPageContainerProps
    extends Omit<UserAddCodeHostsPageProps, 'codeHostExternalServices'> {}

export const UserAddCodeHostsPageContainer: React.FunctionComponent<UserAddCodeHostsPageContainerProps> = props => (
    <UserAddCodeHostsPage
        {...props}
        codeHostExternalServices={{
            github: codeHostExternalServices.github,
            gitlabcom: codeHostExternalServices.gitlabcom,
        }}
    />
)
