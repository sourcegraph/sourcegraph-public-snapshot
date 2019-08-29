import { useEffect, useState } from 'react'
import { Observable } from 'rxjs'
import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { asError, ErrorLike } from '../../../../../../shared/src/util/errors'
import { queryGraphQL } from '../../../../backend/graphql'
import { gitCommitFragment } from '../../../../repo/commits/RepositoryCommitsPage'
import { diffStatFieldsFragment } from '../../../../repo/compare/RepositoryCompareDiffPage'
import { gitRevisionRangeFieldsFragment } from '../../../../repo/compare/RepositoryCompareOverviewPage'

const LOADING: 'loading' = 'loading'

export function useCampaignRepositories(
    campaign: Pick<GQL.ICampaign, 'id'>
): typeof LOADING | GQL.IRepositoryComparison[] | ErrorLike {
    const [data, setData] = useState<typeof LOADING | GQL.IRepositoryComparison[] | ErrorLike>(LOADING)
    useEffect(() => {
        const subscription = queryCampaignRepositories(campaign).subscribe(setData, err => setData(asError(err)))
        return () => subscription.unsubscribe()
    }, [campaign])
    return data
}

function queryCampaignRepositories(campaign: Pick<GQL.ICampaign, 'id'>): Observable<GQL.IRepositoryComparison[]> {
    return queryGraphQL(
        gql`
            query CampaignRepositories($campaign: ID!) {
                node(id: $campaign) {
                    __typename
                    ... on Campaign {
                        repositoryComparisons {
                            baseRepository {
                                id
                                name
                                url
                            }
                            headRepository {
                                id
                                name
                                url
                            }
                            range {
                                ...GitRevisionRangeFields
                            }
                            commits {
                                nodes {
                                    ...GitCommitFields
                                }
                                pageInfo {
                                    hasNextPage
                                }
                            }
                            fileDiffs {
                                diffStat {
                                    ...DiffStatFields
                                }
                            }
                        }
                    }
                }
            }
            ${gitRevisionRangeFieldsFragment}
            ${gitCommitFragment}
            ${diffStatFieldsFragment}
        `,
        { campaign: campaign.id }
    ).pipe(
        map(dataOrThrowErrors),
        map(data => {
            if (!data || !data.node || data.node.__typename !== 'Campaign') {
                throw new Error('campaign not found')
            }
            return data.node.repositoryComparisons
        })
    )
}
