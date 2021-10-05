import React from 'react'
import { UserAddCodeHostsPageProps, UserAddCodeHostsPage } from '../../../user/settings/codeHosts/UserAddCodeHostsPage'

import { codeHostExternalServices } from '../../../components/externalServices/externalServices'
import { Scalars } from '@sourcegraph/shared/src/graphql-operations'

export interface OrgAddCodeHostsPageContainerProps
    extends Omit<UserAddCodeHostsPageProps, 'codeHostExternalServices'> {}

export const OrgAddCodeHostsPageContainer: React.FunctionComponent<OrgAddCodeHostsPageContainerProps> = props => (
    <UserAddCodeHostsPage
        {...props}
        codeHostExternalServices={{
            github: codeHostExternalServices.github,
            gitlabcom: codeHostExternalServices.gitlabcom,
        }}
    />
)
