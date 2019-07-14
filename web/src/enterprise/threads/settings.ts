import * as GQL from '../../../../shared/src/graphql/schema'
import { ChangesetPlan } from '../changesets/plan/plan'

export type PullRequest = {
    repo: string
    label?: string
    items: GQL.ID[]
} & (
    | {
          status: 'pending'
          number: undefined
      }
    | {
          status: 'open' | 'merged' | 'closed'
          title: string
          number: number
          commentsCount: number
          updatedAt: string
          updatedBy: string
      })

export interface PullRequestFields {
    title: string
    branch: string
    description: string
}

export interface CommitStatusRule {
    branch?: string
    infoOnly?: boolean
    enabled?: boolean
}

export interface SlackNotificationRule {
    message?: string
    target?: string
    targetReviewers?: boolean
    remind?: boolean
}

export interface EmailNotificationRule {
    subject?: string
    to?: string
    cc?: string
    digest?: boolean
}

export interface WebhookRule {
    url?: string
}

export interface ChangesetDelta {
    repository: GQL.ID
    base: string
    head: string
}

export interface ChangesetAction {
    user: string
    timestamp: number
    title: string
    detail?: string
}

export interface GitHubPRLink {
    repositoryName: string
    number: number
    url: string

    /**
     * GitHub GraphQL ID of the pull request.
     */
    id: string
}

export interface ThreadSettings {
    queries?: string[]
    pullRequests?: PullRequest[]
    pullRequestTemplate?: Partial<PullRequestFields>
    commitStatusRules?: [CommitStatusRule]
    slackNotificationRules?: [SlackNotificationRule]
    emailNotificationRules?: [EmailNotificationRule]
    webhooks?: [WebhookRule]

    /** @deprecated */
    actions?: { [id: string]: string | undefined }

    /** @deprecated */
    changesetActionDescriptions?: ChangesetAction[]

    plan?: ChangesetPlan
    relatedPRs?: GitHubPRLink[]

    /** @deprecated */
    deltas?: ChangesetDelta[]
}
