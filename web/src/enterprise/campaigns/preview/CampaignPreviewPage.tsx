import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React from 'react'
import { RouteComponentProps } from 'react-router'
import { Link } from 'react-router-dom'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { HeroPage } from '../../../components/HeroPage'
import { GitPullRequestIcon } from '../../../util/octicons'
import { CampaignImpactSummaryBar } from '../common/CampaignImpactSummaryBar'
import { CampaignFileDiffsList } from '../detail/fileDiffs/CampaignFileDiffsList'
import { CampaignRepositoriesList } from '../detail/repositories/CampaignRepositoriesList'
import { CampaignRulesListOLD } from '../detail/rules/CampaignRulesListOLD'
import { useCampaignByID } from '../detail/useCampaignByID'
import { NamespaceCampaignsAreaContext } from '../namespace/NamespaceCampaignsArea'
import { CreateCampaignFromPreviewForm } from './CreateCampaignFromPreviewForm'

interface Props
    extends NamespaceCampaignsAreaContext,
        RouteComponentProps<{ campaignID: string }>,
        PlatformContextProps {}

const LOADING: 'loading' = 'loading'

const CREATE_FORM_EXPANDED_PARAM = 'expand'
const CREATE_FORM_EXPANDED_URL: H.LocationDescriptor = {
    search: new URLSearchParams({ [CREATE_FORM_EXPANDED_PARAM]: '1' }).toString(),
}

/**
 * A page that shows a preview of a campaign created from code actions.
 */
export const CampaignPreviewPage: React.FunctionComponent<Props> = props => {
    const [campaign, onCampaignUpdate] = useCampaignByID(props.match.params.campaignID)
    if (campaign === LOADING) {
        return null // loading
    }
    if (campaign === null) {
        return (
            <HeroPage
                icon={MapSearchIcon}
                title="404: Not Found"
                subtitle="Sorry, the requested campaign was not found."
            />
        )
    }
    if (isErrorLike(campaign)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={campaign.message} />
    }

    const isCreateFormExpanded = new URLSearchParams(props.location.search).get(CREATE_FORM_EXPANDED_PARAM) !== null

    return (
        <div className="campaign-preview-page mt-3 overflow-auto">
            <div className="container">
                <h1 className="mb-3">Preview campaign</h1>
                {isCreateFormExpanded ? (
                    <CreateCampaignFromPreviewForm
                        {...props}
                        campaign={campaign}
                        onCampaignUpdate={onCampaignUpdate}
                        className="border p-3 mb-4"
                        history={props.history}
                    />
                ) : (
                    <div className="alert alert-warning d-flex align-items-center mb-4">
                        <Link to={CREATE_FORM_EXPANDED_URL} className="btn btn-lg btn-success mr-3">
                            <GitPullRequestIcon className="icon-inline mr-1" /> Create campaign
                        </Link>
                        <span className="text-muted">
                            Create branches for this change in all affected repositories and request code reviews.
                        </span>
                    </div>
                )}
                <CampaignImpactSummaryBar {...props} campaign={campaign} />
            </div>
            <hr className="my-4" />
            <div className="container">
                <CampaignRulesListOLD {...props} campaign={campaign} className="mb-4" />
                <CampaignRepositoriesList {...props} campaign={campaign} />
                <CampaignFileDiffsList {...props} campaign={campaign} platformContext={props.platformContext} />
            </div>
        </div>
    )
}
