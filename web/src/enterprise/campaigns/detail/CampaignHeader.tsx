import React from 'react'
import { CampaignsIcon } from '../icons'
import classNames from 'classnames'
import { CampaignFields } from '../../../graphql-operations'
import { Link } from '../../../../../shared/src/components/Link'

interface Props {
    className?: string
    name?: string
    namespace?: Pick<CampaignFields['namespace'], 'namespaceName' | 'url'>
}

/**
 * The header bar for campaigns pages.
 */
export const CampaignHeader: React.FunctionComponent<Props> = ({ className, name, namespace }) => (
    <h1 className={classNames(className)}>
        <CampaignsIcon className="icon-inline mr-2 text-muted" />
        {namespace && (
            <>
                <Link to={namespace.url + '/campaigns'}>{namespace.namespaceName}</Link> /{' '}
            </>
        )}
        {name ?? 'Campaigns'}
        <sup>
            <span className="ml-2 badge badge-merged text-uppercase">Beta</span>
        </sup>
    </h1>
)
