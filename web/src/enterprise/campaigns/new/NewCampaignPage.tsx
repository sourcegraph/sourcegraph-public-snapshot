import React, { useState, useMemo, useCallback } from 'react'
import { Redirect } from 'react-router'
import { map, catchError, concatMap, tap } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import { mutateGraphQL } from '../../../backend/graphql'
import { PageTitle } from '../../../components/PageTitle'
import { Form } from '../../../components/Form'
import { useObservable, useEventObservable } from '../../../util/useObservable'
import { queryNamespaces } from '../../namespaces/backend'
import { asError } from '../../../../../shared/src/util/errors'
import { Observable, concat } from 'rxjs'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { CampaignsIcon } from '../icons'

export const createCampaign = (input: GQL.ICreateCampaignInput): Observable<GQL.ICampaign> =>
    mutateGraphQL(
        gql`
            mutation CreateCampaign($input: CreateCampaignInput!) {
                createCampaign(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => data.createCampaign)
    )

interface Props {}

/**
 * Shows a form to create a new campaign.
 */
export const NewCampaignPage: React.FunctionComponent<Props> = () => {
    const namespaces = useObservable(useMemo(queryNamespaces, []))

    const [namespace, setNamespace] = useState<GQL.ID>()
    const [name, setName] = useState<string>()
    const [description, setDescription] = useState<string>()

    const getNamespace = useCallback((): GQL.ID | undefined => namespace || (namespaces && namespaces[0].id), [
        namespace,
        namespaces,
    ])

    const [onSubmit, createdCampaign] = useEventObservable(
        useCallback(
            (submits: Observable<React.FormEvent<HTMLFormElement>>) =>
                submits.pipe(
                    tap(event => event.preventDefault()),
                    map(() => ({ name: name!, description: description!, namespace: getNamespace()! })),
                    concatMap(input =>
                        concat(['saving' as const], createCampaign(input).pipe(catchError(err => [asError(err)])))
                    )
                ),
            [name, description, getNamespace]
        )
    )

    return (
        <>
            {createdCampaign && createdCampaign !== 'saving' && !(createdCampaign instanceof Error) && (
                <Redirect to={`/campaigns/${createdCampaign.id}`} />
            )}
            <PageTitle title="New campaign" />
            <h2 className="border-bottom pb-2">
                <CampaignsIcon className="icon-inline" /> Create a new campaign
            </h2>
            {createdCampaign instanceof Error && <div className="alert alert-danger">{createdCampaign.message}</div>}
            <Form className="mt-4" onSubmit={onSubmit}>
                <div className="d-flex">
                    <div className="form-group mr-2">
                        <label htmlFor="new-campaign-page__namespace">Owner</label>
                        <select
                            disabled={!namespaces}
                            id="new-campaign-page__namespace"
                            className="form-control"
                            required={true}
                            value={getNamespace()}
                            onChange={event => setNamespace(event.target.value)}
                        >
                            {namespaces &&
                                namespaces.map(namespace => (
                                    <option value={namespace.id} key={namespace.id}>
                                        {namespace.namespaceName}
                                    </option>
                                ))}
                        </select>
                    </div>
                    <div className="form-group">
                        <label htmlFor="new-campaign-page__name">Campaign name</label>
                        <input
                            type="text"
                            id="new-campaign-page__name"
                            className="form-control"
                            required={true}
                            minLength={1}
                            placeholder="Campaign name"
                            value={name || ''}
                            autoFocus={true}
                            onChange={event => setName(event.target.value)}
                        />
                    </div>
                </div>
                <div className="form-group">
                    <label htmlFor="new-campaign-page__description">Campaign description</label>
                    <textarea
                        id="new-campaign-page__description"
                        className="form-control"
                        placeholder="Describe the purpose of this campaign, link to relevant internal documentation, etc."
                        value={description || ''}
                        rows={3}
                        onChange={event => setDescription(event.target.value)}
                    />
                </div>
                <div className="form-group">
                    <button type="submit" className="btn btn-primary">
                        Create campaign
                    </button>
                    {createdCampaign === 'saving' && <LoadingSpinner className="icon-inline" />}
                </div>
            </Form>
        </>
    )
}
