import React from 'react'
import { CampaignsIcon } from '../icons'
import classNames from 'classnames'
import { CampaignFields } from '../../../graphql-operations'
import { Link } from '../../../../../shared/src/components/Link'
import { Timestamp } from '../../../components/time/Timestamp'

interface Props {
    className?: string
    name?: string
    namespace?: Pick<CampaignFields['namespace'], 'namespaceName' | 'url'>
    createdAt?: string
    creator?: CampaignFields['initialApplier'] | null
    verb?: string
}

/**
 * The header bar for campaigns pages.
 */
export const CampaignHeader: React.FunctionComponent<Props> = ({
    className,
    name,
    namespace,
    creator,
    createdAt,
    verb = 'Created',
}) => (
    <div className={classNames(className)}>
        <h1 className="d-inline-block">
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
        {creator !== undefined && createdAt && (
            <span className="text-muted ml-3">
                {verb} <Timestamp date={createdAt} /> by {creator && <Link to={creator.url}>{creator.username}</Link>}
                {!creator && <strong>deleted user</strong>}
            </span>
        )}
    </div>
)
