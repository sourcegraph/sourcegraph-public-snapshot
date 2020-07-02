import * as React from 'react'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { ThemeProps } from '../../../../../shared/src/theme'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { HeroPage } from '../../../components/HeroPage'

interface Props extends ThemeProps {
    campaign: Pick<GQL.ICampaign, 'id' | 'viewerCanAdminister'> & {
        changesets: Pick<GQL.ICampaign['changesets'], 'totalCount'>
    }
    className?: string
}

/**
 * A list of a campaign's changesets changed over a new patch set
 */
export const CampaignUpdateDiff: React.FunctionComponent<Props> = ({ campaign, className }) => {
    if (!campaign.viewerCanAdminister) {
        return <HeroPage body="Updating a campaign is not permitted without campaign admin permissions." />
    }
    return (
        <div className={className}>
            <h3 className="mt-4 mb-2">Preview of changes</h3>
            <div className="alert alert-info mt-2">
                <AlertCircleIcon className="icon-inline" /> You are updating an existing campaign. By clicking 'Update',
                all above changesets that are not 'unmodified' will be updated on the codehost.
            </div>
        </div>
    )
}
