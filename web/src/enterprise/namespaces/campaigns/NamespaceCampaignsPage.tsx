import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import React, { useCallback, useMemo, useState } from 'react'
import { map } from 'rxjs/operators'
import { ExtensionsControllerNotificationProps } from '../../../../../shared/src/extensions/controller'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { pluralize } from '../../../../../shared/src/util/strings'
import { queryGraphQL } from '../../../backend/graphql'
import { NamespaceAreaContext } from '../../../namespaces/NamespaceArea'
import { CampaignRow } from './CampaignRow'
import { NewCampaignForm } from './NewCampaignForm'

const queryNamespaceCampaigns = (namespace: GQL.ID): Promise<GQL.ICampaignConnection> =>
    queryGraphQL(
        gql`
            query NamespaceCampaigns($namespace: ID!) {
                namespace(id: $namespace) {
                    campaigns {
                        nodes {
                            id
                            name
                            url
                        }
                        totalCount
                    }
                }
            }
        `,
        { namespace }
    )
        .pipe(
            map(({ data, errors }) => {
                if (!data || !data.namespace || !data.namespace.campaigns || (errors && errors.length > 0)) {
                    throw createAggregateError(errors)
                }
                return data.namespace.campaigns
            })
        )
        .toPromise()

const LOADING: 'loading' = 'loading'

interface Props extends Pick<NamespaceAreaContext, 'namespace'>, ExtensionsControllerNotificationProps {}

/**
 * Lists a namespace's campaigns.
 */
export const NamespaceCampaignsPage: React.FunctionComponent<Props> = ({ namespace, ...props }) => {
    const [campaignsOrError, setCampaignsOrError] = useState<typeof LOADING | GQL.ICampaignConnection | ErrorLike>(
        LOADING
    )
    const loadCampaigns = useCallback(async () => {
        setCampaignsOrError(LOADING)
        try {
            setCampaignsOrError(await queryNamespaceCampaigns(namespace.id))
        } catch (err) {
            setCampaignsOrError(asError(err))
        }
    }, [namespace])
    // tslint:disable-next-line: no-floating-promises
    useMemo(loadCampaigns, [namespace])

    const [isShowingNewCampaignForm, setIsShowingNewCampaignForm] = useState(false)
    const toggleIsShowingNewCampaignForm = useCallback(() => setIsShowingNewCampaignForm(!isShowingNewCampaignForm), [
        isShowingNewCampaignForm,
    ])

    return (
        <div className="namespace-campaigns-page">
            <div className="d-flex align-items-center justify-content-between mb-3">
                <h2 className="mb-0">Campaigns</h2>
                <button type="button" className="btn btn-success" onClick={toggleIsShowingNewCampaignForm}>
                    New campaign
                </button>
            </div>
            {isShowingNewCampaignForm && (
                <NewCampaignForm
                    namespace={namespace}
                    onDismiss={toggleIsShowingNewCampaignForm}
                    onCampaignCreate={loadCampaigns}
                    className="my-3 p-2 border rounded"
                />
            )}
            {campaignsOrError === LOADING ? (
                <LoadingSpinner className="icon-inline mt-3" />
            ) : isErrorLike(campaignsOrError) ? (
                <div className="alert alert-danger mt-3">{campaignsOrError.message}</div>
            ) : (
                <div className="card">
                    <div className="card-header">
                        <span className="text-muted">
                            {campaignsOrError.totalCount} {pluralize('campaign', campaignsOrError.totalCount)}
                        </span>
                    </div>
                    {campaignsOrError.nodes.length > 0 ? (
                        <ul className="list-group list-group-flush">
                            {campaignsOrError.nodes.map(campaign => (
                                <li key={campaign.id} className="list-group-item p-2">
                                    <CampaignRow {...props} campaign={campaign} onCampaignUpdate={loadCampaigns} />
                                </li>
                            ))}
                        </ul>
                    ) : (
                        <div className="p-2 text-muted">No campaigns yet.</div>
                    )}
                </div>
            )}
        </div>
    )
}
