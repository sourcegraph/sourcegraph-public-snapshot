import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import React, { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ExtensionsControllerProps } from '../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../shared/src/platform/context'
import { isErrorLike } from '../../../../../shared/src/util/errors'
import { ErrorBoundary } from '../../../components/ErrorBoundary'
import { HeroPage } from '../../../components/HeroPage'
import { ThemeProps } from '../../../theme'
import { CampaignsAreaContext } from '../CampaignsArea'

export interface CampaignAreaContext
    extends CampaignsAreaContext,
        ExtensionsControllerProps,
        PlatformContextProps,
        ThemeProps {
    /** The rule ID. */
    ruleID: GQL.ID

    /** The rule, queried from the GraphQL API. */
    rule: GQL.ICampaign

    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
}

interface Props extends Pick<CampaignAreaContext, Exclude<keyof CampaignAreaContext, 'rule'>> {}

const LOADING = 'loading' as const

/**
 * The area for a single rule.
 */
export const CampaignArea: React.FunctionComponent<Props> = ({ ruleID, setBreadcrumbItem, ...props }) => {
    const ruleOrError = useCampaignByID(ruleID)

    useEffect(() => {
        setBreadcrumbItem(
            ruleOrError !== LOADING && ruleOrError !== null && !isErrorLike(ruleOrError)
                ? { text: ruleOrError.name, to: ruleOrError.url }
                : undefined
        )
        return () => setBreadcrumbItem(undefined)
    }, [ruleOrError, setBreadcrumbItem])

    if (ruleOrError === LOADING) {
        return <LoadingSpinner className="icon-inline mx-auto my-4" />
    }
    if (ruleOrError === null) {
        return <HeroPage icon={AlertCircleIcon} title="Campaign not found" />
    }
    if (isErrorLike(ruleOrError)) {
        return <HeroPage icon={AlertCircleIcon} title="Error" subtitle={ruleOrError.message} />
    }

    const context: CampaignAreaContext = {
        ...props,
        ruleID,
        rule: ruleOrError,
        setBreadcrumbItem,
    }

    return (
        <div className="rule-area flex-1">
            <div className="d-flex align-items-center justify-content-between border-top border-bottom py-3 my-3">
                <div className="d-flex align-items-center">
                    <div className="badge border border-success text-success font-size-base py-2 px-3 mr-3">Active</div>
                    Last action 11 minutes ago
                </div>
                <div>
                    <Link className="btn btn-secondary mr-2" to="#TODO">
                        Deactivate
                    </Link>
                    <Link className="btn btn-secondary mr-2" to="#edit">
                        Delete
                    </Link>
                </div>
            </div>
            <h2>{ruleOrError.name}</h2>
            <div className="flex-1 d-flex flex-column overflow-auto">
                <ErrorBoundary location={props.location}>RULE AREA</ErrorBoundary>
            </div>
        </div>
    )
}
