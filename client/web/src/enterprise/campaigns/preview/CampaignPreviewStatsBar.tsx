import React from 'react'
import { DiffStat } from '../../../components/diff/DiffStat'
import { CampaignSpecFields } from '../../../graphql-operations'
import { ChangesetAddedIcon, ChangesetModifiedIcon, ChangesetRemovedIcon } from './icons'
import {
    PreviewActionClose,
    PreviewActionImport,
    PreviewActionPublish,
    PreviewActionReopen,
    PreviewActionUndraft,
    PreviewActionUpdate,
} from './list/PreviewActions'

const actionClassNames =
    'd-flex flex-column justify-content-center align-items-center campaign-preview-stats-bar__stat mx-2 campaign-preview-stats-bar__state'

export interface CampaignPreviewStatsBarProps {
    campaignSpec: CampaignSpecFields
}

export const CampaignPreviewStatsBar: React.FunctionComponent<CampaignPreviewStatsBarProps> = ({ campaignSpec }) => (
    <div className="d-flex flex-wrap mb-3 align-items-center">
        <h2 className="m-0 align-self-center">
            <span className="badge badge-info text-uppercase mb-0">Preview</span>
        </h2>
        <div className="campaign-preview-stats-bar__divider d-none d-sm-block mx-3" />
        <DiffStat
            {...campaignSpec.diffStat}
            separateLines={true}
            expandedCounts={true}
            className="campaign-preview-stats-bar__diff"
        />
        <div className="campaign-preview-stats-bar__horizontal-divider d-block d-sm-none my-3" />
        <div className="campaign-preview-stats-bar__divider mx-3 d-none d-sm-block d-md-none" />
        <div className="flex-grow-1 d-flex justify-content-end campaign-preview-stats-bar__metrics">
            <PreviewStatsAdded count={campaignSpec.applyPreview.stats.added} />
            <PreviewStatsRemoved count={campaignSpec.applyPreview.stats.removed} />
            <PreviewStatsModified count={campaignSpec.applyPreview.stats.modified} />
        </div>
        <div className="campaign-preview-stats-bar__horizontal-divider d-block d-md-none my-3" />
        <div className="campaign-preview-stats-bar__divider d-none d-md-block ml-3 mr-2" />
        <div className="campaign-preview-stats-bar__states d-flex justify-content-end">
            <PreviewActionReopen
                className={actionClassNames}
                label={`${campaignSpec.applyPreview.stats.reopen} reopen`}
            />
            <PreviewActionClose
                className={actionClassNames}
                label={`${campaignSpec.applyPreview.stats.reopen} close`}
            />
            <PreviewActionUpdate
                className={actionClassNames}
                label={`${campaignSpec.applyPreview.stats.update} update`}
            />
            <PreviewActionUndraft
                className={actionClassNames}
                label={`${campaignSpec.applyPreview.stats.undraft} undraft`}
            />
            <PreviewActionPublish
                className={actionClassNames}
                label={`${
                    campaignSpec.applyPreview.stats.publish + campaignSpec.applyPreview.stats.publishDraft
                } publish`}
            />
            <PreviewActionImport
                className={actionClassNames}
                label={`${campaignSpec.applyPreview.stats.import} import`}
            />
        </div>
    </div>
)

export const PreviewStatsAdded: React.FunctionComponent<{ count: number }> = ({ count }) => (
    <div className="d-flex flex-column campaign-preview-stats-bar__stat mr-2 text-nowrap">
        <div className="d-flex flex-column align-items-center justify-content-center">
            <span className="preview-stats-added__line">&nbsp;</span>
            <span className="preview-stats-added__icon d-flex justify-content-center align-items-center">
                <ChangesetAddedIcon />
            </span>
            <span className="preview-stats-added__line">&nbsp;</span>
        </div>
        {count} added
    </div>
)
export const PreviewStatsModified: React.FunctionComponent<{ count: number }> = ({ count }) => (
    <div className="d-flex flex-column campaign-preview-stats-bar__stat text-nowrap ml-2">
        <div className="d-flex flex-column align-items-center">
            <span className="preview-stats-modified__line">&nbsp;</span>
            <span className="preview-stats-modified__icon d-flex justify-content-center align-items-center">
                <ChangesetModifiedIcon />
            </span>
            <span className="preview-stats-modified__line">&nbsp;</span>
        </div>
        {count} modified
    </div>
)
export const PreviewStatsRemoved: React.FunctionComponent<{ count: number }> = ({ count }) => (
    <div className="d-flex flex-column campaign-preview-stats-bar__stat mx-2 text-nowrap">
        <div className="d-flex flex-column align-items-center">
            <span className="preview-stats-removed__line">&nbsp;</span>
            <span className="preview-stats-removed__icon d-flex justify-content-center align-items-center">
                <ChangesetRemovedIcon />
            </span>
            <span className="preview-stats-removed__line">&nbsp;</span>
        </div>
        {count} removed
    </div>
)
