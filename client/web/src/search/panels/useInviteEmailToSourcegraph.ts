import { gql, MutationFunctionOptions, FetchResult } from '@apollo/client'

import { useMutation } from '@sourcegraph/http-client'

import { Exact, InviteEmailToSourcegraphResult, InviteEmailToSourcegraphVariables } from '../../graphql-operations'

const INVITE_EMAIL_TO_SOURCEGRAPH = gql`
    mutation InviteEmailToSourcegraph($email: String!) {
        inviteEmailToSourcegraph(email: $email) {
            alwaysNil
        }
    }
`

type UseInviteEmailToSourcegraphResult = (
    options?:
        | MutationFunctionOptions<
              InviteEmailToSourcegraphResult,
              Exact<{
                  email: string
              }>
          >
        | undefined
) => Promise<FetchResult<InviteEmailToSourcegraphResult>>

export const useInviteEmailToSourcegraph = (): UseInviteEmailToSourcegraphResult => {
    const [inviteEmailToSourcegraph] = useMutation<InviteEmailToSourcegraphResult, InviteEmailToSourcegraphVariables>(
        INVITE_EMAIL_TO_SOURCEGRAPH
    )
    return inviteEmailToSourcegraph
}
