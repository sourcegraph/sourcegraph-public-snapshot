import { Maybe } from '@sourcegraph/shared/src/graphql-operations'

export interface InvitableCollaborator {
    email: string
    displayName: string
    name: string
    avatarURL: Maybe<string>
}
