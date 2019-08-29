import React from 'react'
import { CommitStatusRuleActionType } from './types/CommitStatusRuleActionType'
import { SlackRuleActionType } from './types/SlackRuleActionType'
import { WebhookRuleActionType } from './types/WebhookRuleActionType'
import { EmailRuleActionType } from './types/EmailRuleActionType'
import { EditorRuleActionType } from './types/EditorRuleActionType'
import { BranchReviewRuleActionType } from './types/BranchReviewRuleActionType'

export interface RuleActionTypeComponentContext<T extends object> {
    value: T
    onChange: (value: T) => void
}

export interface RuleActionType<ID extends string, T extends { type: ID }> {
    id: ID
    title: string
    icon?: React.ComponentType<{ className?: string }>
    renderForm: React.FunctionComponent<RuleActionTypeComponentContext<T>>
    initialValue: T
}

export const RULE_ACTION_TYPES = [
    BranchReviewRuleActionType,
    CommitStatusRuleActionType,
    EmailRuleActionType,
    SlackRuleActionType,
    WebhookRuleActionType,
    EditorRuleActionType,
]

/**
 * A generic rule action value type.
 */
export type GenericRuleAction = { type: string } & object

/**
 * The union of all rule action value types.
 */
export type RuleAction = (typeof RULE_ACTION_TYPES)[number]['initialValue']
