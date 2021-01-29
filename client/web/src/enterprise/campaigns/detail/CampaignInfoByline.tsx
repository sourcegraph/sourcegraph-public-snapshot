import React from 'react'
import { CampaignFields } from '../../../graphql-operations'
import { Link } from '../../../../../shared/src/components/Link'
import { Timestamp } from '../../../components/time/Timestamp'

interface Props extends Pick<CampaignFields, 'createdAt' | 'initialApplier' | 'lastAppliedAt' | 'lastApplier'> {}

/**
 * The created/updated byline in the campaign header.
 */
export const CampaignInfoByline: React.FunctionComponent<Props> = ({
    createdAt,
    initialApplier,
    lastAppliedAt,
    lastApplier,
}) => (
    <>
        {initialApplier ? <Link to={initialApplier.url}>{initialApplier.username}</Link> : 'A deleted user'}{' '}
        <Timestamp date={createdAt} />
        {lastAppliedAt !== createdAt && (
            <>
                <span className="mx-2">•</span>
                {lastApplier?.username !== initialApplier?.username && (
                    <>{lastApplier ? <Link to={lastApplier.url}>{lastApplier.username}</Link> : 'A deleted user'} </>
                )}
                updated <Timestamp date={lastAppliedAt} />
            </>
        )}
    </>
)
