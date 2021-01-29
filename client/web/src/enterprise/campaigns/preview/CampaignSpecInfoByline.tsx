import React from 'react'
import { CampaignSpecFields } from '../../../graphql-operations'
import { Link } from '../../../../../shared/src/components/Link'
import { Timestamp } from '../../../components/time/Timestamp'

interface Props extends Pick<CampaignSpecFields, 'createdAt' | 'creator'> {}

/**
 * The uploaded at byline to the campaign apply page header.
 */
export const CampaignSpecInfoByline: React.FunctionComponent<Props> = ({ createdAt, creator }) => (
    <>
        Uploaded <Timestamp date={createdAt} /> by {creator && <Link to={creator.url}>{creator.username}</Link>}
        {!creator && <strong>deleted user</strong>}
    </>
)
