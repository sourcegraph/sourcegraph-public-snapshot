import * as H from 'history'
import React, { useEffect, useMemo } from 'react'
import { useObservable } from '../../../../../shared/src/util/useObservable'
import { PageTitle } from '../../../components/PageTitle'
import {
    fetchCampaignSpecById as _fetchCampaignSpecById,
    queryChangesetSpecs,
    queryChangesetSpecFileDiffs,
} from './backend'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { CampaignHeader } from '../detail/CampaignHeader'
import { ChangesetSpecList } from './ChangesetSpecList'
import { ThemeProps } from '../../../../../shared/src/theme'
import { CreateUpdateCampaignAlert } from './CreateUpdateCampaignAlert'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import { HeroPage } from '../../../components/HeroPage'
import { CampaignDescription } from '../detail/CampaignDescription'
import { CampaignSpecInfoByline } from './CampaignSpecInfoByline'
import { TelemetryProps } from '../../../../../shared/src/telemetry/telemetryService'
import { Link } from '../../../../../shared/src/components/Link'
import { AuthenticatedUser } from '../../../auth'
import { ExternalServiceKind } from '../../../graphql-operations'
import { defaultExternalServices } from '../../../components/externalServices/externalServices'
import { pluralize } from '../../../../../shared/src/util/strings'

export interface CampaignApplyPageProps extends ThemeProps, TelemetryProps {
    specID: string
    history: H.History
    location: H.Location
    authenticatedUser: Pick<AuthenticatedUser, 'url'>

    /** Used for testing. */
    fetchCampaignSpecById?: typeof _fetchCampaignSpecById
    /** Used for testing. */
    queryChangesetSpecs?: typeof queryChangesetSpecs
    /** Used for testing. */
    queryChangesetSpecFileDiffs?: typeof queryChangesetSpecFileDiffs
    /** Expand changeset descriptions, for testing only. */
    expandChangesetDescriptions?: boolean
}

export const CampaignApplyPage: React.FunctionComponent<CampaignApplyPageProps> = ({
    specID,
    history,
    location,
    authenticatedUser,
    isLightTheme,
    telemetryService,
    fetchCampaignSpecById = _fetchCampaignSpecById,
    queryChangesetSpecs,
    queryChangesetSpecFileDiffs,
    expandChangesetDescriptions,
}) => {
    const spec = useObservable(useMemo(() => fetchCampaignSpecById(specID), [specID, fetchCampaignSpecById]))

    useEffect(() => {
        telemetryService.logViewEvent('CampaignApplyPage')
    }, [telemetryService])

    if (spec === undefined) {
        return (
            <div className="text-center">
                <LoadingSpinner className="icon-inline mx-auto my-4" />
            </div>
        )
    }
    if (spec === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign spec not found" />
    }

    return (
        <>
            <PageTitle title="Apply campaign spec" />
            <CampaignHeader
                name={spec.description.name}
                namespace={spec.namespace}
                className="test-campaign-apply-page"
            />
            <CampaignSpecInfoByline createdAt={spec.createdAt} creator={spec.creator} className="mb-3" />
            {spec.viewerMissingCodeHostCredentials.totalCount > 0 && (
                <div className="alert alert-warning">
                    <p className="alert-title">
                        You don't have credentials configured for{' '}
                        {pluralize(
                            'this code host',
                            spec.viewerMissingCodeHostCredentials.totalCount,
                            'these code hosts'
                        )}
                    </p>
                    <ul>
                        {spec.viewerMissingCodeHostCredentials.nodes.map(node => (
                            <MissingCodeHost {...node} key={node.externalServiceKind + node.externalServiceURL} />
                        ))}
                    </ul>
                    <p className="mb-0">
                        Configure {pluralize('it', spec.viewerMissingCodeHostCredentials.totalCount, 'them')} in your{' '}
                        <Link to={`${authenticatedUser.url}/settings/campaigns`} target="_blank" rel="noopener">
                            campaigns user settings
                        </Link>{' '}
                        to apply this spec.
                    </p>
                </div>
            )}
            <CreateUpdateCampaignAlert
                history={history}
                specID={spec.id}
                campaign={spec.appliesToCampaign}
                viewerCanAdminister={spec.viewerCanAdminister}
                telemetryService={telemetryService}
            />
            <CampaignDescription history={history} description={spec.description.description} />
            <ChangesetSpecList
                campaignSpecID={specID}
                history={history}
                location={location}
                isLightTheme={isLightTheme}
                queryChangesetSpecs={queryChangesetSpecs}
                queryChangesetSpecFileDiffs={queryChangesetSpecFileDiffs}
                expandChangesetDescriptions={expandChangesetDescriptions}
            />
        </>
    )
}

const MissingCodeHost: React.FunctionComponent<{
    externalServiceKind: ExternalServiceKind
    externalServiceURL: string
}> = ({ externalServiceKind, externalServiceURL }) => {
    const Icon = defaultExternalServices[externalServiceKind].icon
    return (
        <li>
            <Icon className="icon-inline mr-2" />
            {externalServiceURL}
        </li>
    )
}
