import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import React, { useMemo } from 'react'
import { highlightCodeSafe } from '../../../../../shared/src/util/markdown'
import { CampaignFields } from '../../../graphql-operations'

export interface CampaignSpecTabProps {
    campaignName: CampaignFields['name']
    originalInput: CampaignFields['currentSpec']['originalInput']
}

export const CampaignSpecTab: React.FunctionComponent<CampaignSpecTabProps> = ({ originalInput, campaignName }) => {
    const downloadUrl = useMemo(() => 'data:text/plain;charset=utf-8,' + encodeURIComponent(originalInput), [
        originalInput,
    ])
    const highlightedInput = useMemo(() => ({ __html: highlightCodeSafe(originalInput, 'yaml') }), [originalInput])
    return (
        <div className="mt-4">
            <div className="d-flex justify-content-between align-items-center mb-2 test-campaigns-spec">
                <p className="m-0">This campaign was created by applying the following campaign spec:</p>
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
