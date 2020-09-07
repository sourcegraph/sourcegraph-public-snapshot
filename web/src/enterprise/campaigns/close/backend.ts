import { CloseCampaignResult, CloseCampaignVariables } from '../../../graphql-operations'
import { requestGraphQL, gql, dataOrThrowErrors } from '../../../../../shared/src/graphql/graphql'

export async function closeCampaign({ campaign, closeChangesets }: CloseCampaignVariables): Promise<void> {
    const result = await requestGraphQL<CloseCampaignResult, CloseCampaignVariables>({
        request: gql`
            mutation CloseCampaign($campaign: ID!, $closeChangesets: Boolean) {
                closeCampaign(campaign: $campaign, closeChangesets: $closeChangesets) {
                    id
                }
            }
        `,
        variables: { campaign, closeChangesets },
    }).toPromise()
    dataOrThrowErrors(result)
}
