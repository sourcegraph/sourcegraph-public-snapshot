import React from 'react'
import { CampaignsIcon } from '../icons'
import { CampaignFields } from '../../../graphql-operations'
import { Link } from '../../../../../shared/src/components/Link'
import { PageHeader } from '../../../components/PageHeader'

interface Props {
    className?: string
    name: string
    namespace: Pick<CampaignFields['namespace'], 'namespaceName' | 'url'>
    actionSection?: JSX.Element
}

/**
 * The header bar for campaigns pages.
 */
export const CampaignHeader: React.FunctionComponent<Props> = ({ className, name, namespace, actionSection }) => (
    <PageHeader
        icon={CampaignsIcon}
        title={
            <>
                <Link to={namespace.url + '/campaigns'}>{namespace.namespaceName}</Link> / {name}{' '}
                <sup>
                    <span className="badge badge-merged text-uppercase">Beta</span>
                </sup>
            </>
        }
        actions={actionSection}
        className={className}
    />
)
