import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import React, { useMemo } from 'react'
import { Link } from '../../../../../shared/src/components/Link'
import { highlightCodeSafe } from '../../../../../shared/src/util/markdown'
import { Timestamp } from '../../../components/time/Timestamp'
import { CampaignFields } from '../../../graphql-operations'

export interface CampaignSpecTabProps {
    campaign: Pick<CampaignFields, 'name' | 'createdAt' | 'lastApplier' | 'lastAppliedAt'>
    originalInput: CampaignFields['currentSpec']['originalInput']
}

export const CampaignSpecTab: React.FunctionComponent<CampaignSpecTabProps> = ({
    campaign: { name: campaignName, createdAt, lastApplier, lastAppliedAt },
    originalInput,
}) => {
    const downloadUrl = useMemo(() => 'data:text/plain;charset=utf-8,' + encodeURIComponent(originalInput), [
        originalInput,
    ])
    const highlightedInput = useMemo(() => ({ __html: highlightCodeSafe(originalInput, 'yaml') }), [originalInput])
    return (
        <div className="mt-4">
            <div className="d-flex justify-content-between align-items-center mb-2 test-campaigns-spec">
                <p className="m-0">
                    {lastApplier ? <Link to={lastApplier.url}>{lastApplier.username}</Link> : 'A deleted user'}{' '}
                    {createdAt === lastAppliedAt ? 'created' : 'updated'} this campaign{' '}
                    <Timestamp date={lastAppliedAt} /> by applying the following campaign spec:
                </p>
                <a
                    download={`${campaignName}.campaign.yaml`}
                    href={downloadUrl}
                    className="text-right btn btn-secondary text-nowrap"
                    data-tooltip={`Download ${campaignName}.campaign.yaml`}
                >
                    <FileDownloadIcon className="icon-inline" /> Download YAML
                </a>
            </div>
            <div className="mb-3">
                <div className="campaign-spec-tab__specfile rounded p-3">
                    <pre className="m-0" dangerouslySetInnerHTML={highlightedInput} />
                </div>
            </div>
        </div>
    )
}
