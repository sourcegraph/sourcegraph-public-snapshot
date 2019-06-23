import * as GQL from '../../../../shared/src/graphql/schema'
import { FileDiff } from './detail/changes/computeDiff'

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

export interface ThreadSettings {
    providers?: string[]
    queries?: string[]
    pullRequests?: PullRequest[]
    pullRequestTemplate?: Partial<PullRequestFields>
    commitStatusRules?: [CommitStatusRule]
    slackNotificationRules?: [SlackNotificationRule]
    emailNotificationRules?: [EmailNotificationRule]
    webhooks?: [WebhookRule]
    actions?: { [id: string]: string | undefined }

    previewChangesetDiff?: FileDiff[]
}
