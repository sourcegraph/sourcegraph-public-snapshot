import { gql } from '@apollo/client'

export const ORG_MEMBERS_QUERY = gql`
    query OrganizationMembers($id: ID!) {
        node(id: $id) {
            ... on Org {
                viewerCanAdminister
                members {
                    nodes {
                        id
                        username
                        displayName
                        avatarURL
                    }
                    totalCount
                }
            }
        }
    }
`

export const SEARCH_USERS_AUTOCOMPLETE_QUERY = gql`
    query AutocompleteMembersSearch($organization: ID!, $query: String!) {
        autocompleteMembersSearch(organization: $organization, query: $query) {
            id
            username
            displayName
            avatarURL
            inOrg
        }
    }
`

export const ORG_MEMBER_REMOVE_MUTATION = gql`
    mutation RemoveUserFromOrg($user: ID!, $organization: ID!) {
        removeUserFromOrganization(user: $user, organization: $organization) {
            alwaysNil
        }
    }
`

export const INVITE_USERNAME_OR_EMAIL_TO_ORG_MUTATION = gql`
    mutation InviteUserToOrg($organization: ID!, $username: String, $email: String) {
        inviteUserToOrganization(organization: $organization, username: $username, email: $email) {
            ...InviteUserToOrganizationFields
        }
    }

    fragment InviteUserToOrganizationFields on InviteUserToOrganizationResult {
        sentInvitationEmail
        invitationURL
    }
`

export const ADD_USERNAME_OR_EMAIL_TO_ORG_MUTATION = gql`
    mutation AddUserToOrganization($organization: ID!, $username: String!) {
        addUserToOrganization(organization: $organization, username: $username) {
            alwaysNil
        }
    }
`

export const ORG_PENDING_INVITES_QUERY = gql`
    query PendingInvitations($id: ID!) {
        pendingInvitations(organization: $id) {
            id
            recipientEmail
            expiresAt
            respondURL
            recipient {
                id
                username
                displayName
                avatarURL
            }
            revokedAt
            sender {
                id
                displayName
                username
            }
            organization {
                name
            }
            createdAt
            notifiedAt
        }
    }
`
export const ORG_REVOKE_INVITATION_MUTATION = gql`
    mutation RevokeInvite($id: ID!) {
        revokeOrganizationInvitation(organizationInvitation: $id) {
            alwaysNil
        }
    }
`

export const ORG_RESEND_INVITATION_MUTATION = gql`
    mutation ResendOrgInvitation($id: ID!) {
        resendOrganizationInvitationNotification(organizationInvitation: $id) {
            alwaysNil
        }
    }
`
