import { map } from 'rxjs/operators'
import { dataOrThrowErrors, gql } from '../../../../../shared/src/graphql/graphql'
import { queryGraphQL, mutateGraphQL } from '../../../backend/graphql'
import { Observable, of } from 'rxjs'
import { ID, ICampaign, IUpdateCampaignInput, ICreateCampaignInput } from '../../../../../shared/src/graphql/schema'

const campaignFragment = gql`
    fragment CampaignFields on Campaign {
        id
        namespace {
            id
            namespaceName
        }
        author {
            username
            avatarURL
        }
        name
        description
        createdAt
        updatedAt
        url
        changesets {
            nodes {
                id
                title
                body
                state
                reviewState
                repository {
                    name
                    url
                }
                externalURL {
                    url
                }
                createdAt
            }
        }
        # TODO move to separate query and configure from/to
        changesetCountsOverTime {
            date
            merged
            closed
            openApproved
            openChangesRequested
            openPending
        }
    }
`

export async function updateCampaign(update: IUpdateCampaignInput): Promise<ICampaign> {
    const result = await mutateGraphQL(
        gql`
            mutation UpdateCampaign($update: UpdateCampaignInput!) {
                updateCampaign(input: $update) {
                    ...CampaignFields
                }
            }
            ${campaignFragment}
        `,
        { update }
    ).toPromise()
    return dataOrThrowErrors(result).updateCampaign
}

export async function createCampaign(input: ICreateCampaignInput): Promise<ICampaign> {
    const result = await mutateGraphQL(
        gql`
            mutation CreateCampaign($input: CreateCampaignInput!) {
                createCampaign(input: $input) {
                    id
                    url
                }
            }
        `,
        { input }
    ).toPromise()
    return dataOrThrowErrors(result).createCampaign
}

export async function deleteCampaign(campaign: ID): Promise<void> {
    const result = await mutateGraphQL(
        gql`
            mutation DeleteCampaign($campaign: ID!) {
                deleteCampaign(campaign: $campaign) {
                    alwaysNil
                }
            }
        `,
        { campaign }
    ).toPromise()
    dataOrThrowErrors(result)
}

const mockData = false
export const fetchCampaignById = (campaign: ID): Observable<ICampaign | null> =>
    mockData
        ? of(({
              __typename: 'Campaign',
              id: 'Q2FtcGFpZ246MQ==',
              namespace: { id: 'VXNlcjox', namespaceName: 'felix' },
              author: { username: 'felix', avatarURL: null },
              name: 'test camp',
              description: 'asdasd',
              createdAt: '2019-09-12T20:18:10Z',
              updatedAt: '2019-10-08T16:39:48Z',
              url: '/users/felix/campaigns/Q2FtcGFpZ246MQ==',
              changesets: { nodes: [] },
          } as unknown) as ICampaign)
        : queryGraphQL(
              gql`
                  query CampaignByID($campaign: ID!) {
                      node(id: $campaign) {
                          __typename
                          ... on Campaign {
                              ...CampaignFields
                          }
                      }
                  }
                  ${campaignFragment}
              `,
              { campaign }
          ).pipe(
              map(dataOrThrowErrors),
              map(({ node }) => {
                  if (!node) {
                      return null
                  }
                  if (node.__typename !== 'Campaign') {
                      throw new Error(`The given ID is a ${node.__typename}, not a Campaign`)
                  }
                  return node
              })
          )
