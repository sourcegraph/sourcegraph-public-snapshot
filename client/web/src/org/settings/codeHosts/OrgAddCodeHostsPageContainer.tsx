import React from 'react'

import { codeHostExternalServices } from '../../../components/externalServices/externalServices'
import { UserAddCodeHostsPageProps, UserAddCodeHostsPage } from '../../../user/settings/codeHosts/UserAddCodeHostsPage'

export interface OrgAddCodeHostsPageContainerProps
    extends Omit<UserAddCodeHostsPageProps, 'codeHostExternalServices'> {}

export const OrgAddCodeHostsPageContainer: React.FunctionComponent<
    React.PropsWithChildren<OrgAddCodeHostsPageContainerProps>
> = props => (
    <UserAddCodeHostsPage
        {...props}
        codeHostExternalServices={{
            github: codeHostExternalServices.github,
            gitlabcom: codeHostExternalServices.gitlabcom,
        }}
    />
)
