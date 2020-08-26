import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import React, { useMemo } from 'react'
import { CampaignFields } from '../../../graphql-operations'

export interface CampaignSpecTabProps {
    originalInput: CampaignFields['currentSpec']['originalInput']
}

export const CampaignSpecTab: React.FunctionComponent<CampaignSpecTabProps> = ({ originalInput }) => {
    const downloadUrl = useMemo(() => 'data:text/plain;charset=utf-8,' + encodeURIComponent(originalInput), [
        originalInput,
    ])
    return (
        <>
            <div className="d-flex justify-content-between align-items-center mb-2">
                <p className="m-0">This campaign was created from the folowing spec:</p>
                <a
                    download="campaign-spec.yaml"
                    href={downloadUrl}
                    className="text-right btn btn-secondary text-nowrap"
                    data-tooltip="Download campaign-spec.yaml"
                >
                    <FileDownloadIcon className="icon-inline" /> Download yaml
                </a>
            </div>
            <div className="bg-light rounded p-2 mb-3 col-lg-8 offset-lg-2">
                <pre className="m-0">{originalInput}</pre>
            </div>
        </>
    )
}
