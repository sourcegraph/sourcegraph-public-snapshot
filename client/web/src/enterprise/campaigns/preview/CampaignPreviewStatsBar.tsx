import React from 'react'
import { DiffStat } from '../../../components/diff/DiffStat'
import { CampaignSpecFields } from '../../../graphql-operations'
import {
    PreviewActionClose,
    PreviewActionImport,
    PreviewActionPublish,
    PreviewActionReopen,
    PreviewActionUndraft,
    PreviewActionUpdate,
} from './list/PreviewAction'

export interface CampaignPreviewStatsBarProps {
    campaignSpec: CampaignSpecFields
}

export const CampaignPreviewStatsBar: React.FunctionComponent<CampaignPreviewStatsBarProps> = ({ campaignSpec }) => (
    <div className="d-flex mb-3">
        <h2 className="m-0 mr-3 create-update-campaign-alert__badge">
            <span className="badge badge-info text-uppercase mb-0">Preview</span>
        </h2>
        <DiffStat {...campaignSpec.diffStat} separateLines={true} expandedCounts={true} />
        <div className="flex-grow-1 d-flex justify-content-between">
            <PreviewStatsAdded count={campaignSpec.applyPreview.stats.added} />
            <PreviewStatsRemoved count={campaignSpec.applyPreview.stats.removed} />
            <PreviewStatsModified count={campaignSpec.applyPreview.stats.modified} />
        </div>
        <div className="flex-grow-1 d-flex justify-content-between">
            <PreviewActionReopen label={`${campaignSpec.applyPreview.stats.reopen} reopen`} />
            <PreviewActionClose label={`${campaignSpec.applyPreview.stats.reopen} close`} />
            <PreviewActionUpdate label={`${campaignSpec.applyPreview.stats.update} update`} />
            <PreviewActionUndraft label={`${campaignSpec.applyPreview.stats.undraft} undraft`} />
            {/* <PreviewActionUnpublished label={`${campaignSpec.applyPreview.stats.reopen} close`} /> */}
            <PreviewActionPublish
                label={`${
                    campaignSpec.applyPreview.stats.publish + campaignSpec.applyPreview.stats.publishDraft
                } publish`}
            />
            <PreviewActionImport label={`${campaignSpec.applyPreview.stats.import} import`} />
        </div>
    </div>
)

export const PreviewStatsAdded: React.FunctionComponent<{ count: number }> = ({ count }) => (
    <div className="d-flex flex-column">
        <div className="d-flex flex-column align-items-center">
            <span style={{ width: '0.25rem', height: '0.25rem', backgroundColor: 'var(--oc-lime-6)' }}>&nbsp;</span>
            <span style={{ color: 'var(--oc-lime-9)' }}>+</span>
            <span style={{ width: '0.25rem', height: '0.25rem', backgroundColor: 'var(--oc-lime-6)' }}>&nbsp;</span>
        </div>
        {count} added
    </div>
)
export const PreviewStatsModified: React.FunctionComponent<{ count: number }> = ({ count }) => (
    <div className="d-flex flex-column">
        <div className="d-flex flex-column align-items-center">
            <span style={{ width: '0.25rem', height: '0.25rem', backgroundColor: 'var(--oc-yellow-5)' }}>&nbsp;</span>
            <span style={{ color: 'var(--oc-orange-9)' }}>&#9679;</span>
            <span style={{ width: '0.25rem', height: '0.25rem', backgroundColor: 'var(--oc-yellow-5)' }}>&nbsp;</span>
        </div>
        {count} modified
    </div>
)
export const PreviewStatsRemoved: React.FunctionComponent<{ count: number }> = ({ count }) => (
    <div className="d-flex flex-column">
        <div className="d-flex flex-column align-items-center">
            <span style={{ width: '0.25rem', height: '0.25rem', backgroundColor: 'var(--danger)' }}>&nbsp;</span>
            <span style={{ color: 'var(--oc-red-9)' }}>-</span>
            <span style={{ width: '0.25rem', height: '0.25rem', backgroundColor: 'var(--danger)' }}>&nbsp;</span>
        </div>
        {count} removed
    </div>
)
