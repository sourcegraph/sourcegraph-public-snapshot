import type { AuthenticatedUser } from '@sourcegraph/shared/src/auth'

import {
    type CodeMonitorFields,
    type MonitorActionInput,
    type MonitorEditActionInput,
    type MonitorEmailInput,
    MonitorEmailPriority,
    type MonitorWebhookInput,
    type MonitorSlackWebhookInput,
    type MonitorWebhookFields,
    type MonitorSlackWebhookFields,
    type MonitorEmailFields,
} from '../../graphql-operations'

function convertEmailAction(
    action: MonitorEmailFields,
    authenticatedUserId: AuthenticatedUser['id']
): MonitorEmailInput {
    return {
        enabled: action.enabled,
        includeResults: action.includeResults,
        priority: MonitorEmailPriority.NORMAL,
        recipients: [authenticatedUserId],
        header: '',
    }
}

function convertSlackWebhookAction(action: MonitorSlackWebhookFields): MonitorSlackWebhookInput {
    return {
        enabled: action.enabled,
        includeResults: action.includeResults,
        url: action.url,
    }
}

function convertWebhookAction(action: MonitorWebhookFields): MonitorWebhookInput {
    return {
        enabled: action.enabled,
        includeResults: action.includeResults,
        url: action.url,
    }
}

export function convertActionsForCreate(
    actions: CodeMonitorFields['actions']['nodes'],
    authenticatedUserId: AuthenticatedUser['id']
): MonitorActionInput[] {
    return actions.map(action => {
        switch (action.__typename) {
            case 'MonitorEmail': {
                return {
                    email: convertEmailAction(action, authenticatedUserId),
                }
            }
            case 'MonitorSlackWebhook': {
                return {
                    slackWebhook: convertSlackWebhookAction(action),
                }
            }
            case 'MonitorWebhook': {
                return {
                    webhook: convertWebhookAction(action),
                }
            }
        }
    })
}

export function convertActionsForUpdate(
    actions: CodeMonitorFields['actions']['nodes'],
    authenticatedUserId: AuthenticatedUser['id']
): MonitorEditActionInput[] {
    return actions.map(action => {
        // Convert empty IDs to null so action is created
        switch (action.__typename) {
            case 'MonitorEmail': {
                return {
                    email: {
                        id: action.id || null,
                        update: convertEmailAction(action, authenticatedUserId),
                    },
                }
            }
            case 'MonitorSlackWebhook': {
                return {
                    slackWebhook: {
                        id: action.id || null,
                        update: convertSlackWebhookAction(action),
                    },
                }
            }
            case 'MonitorWebhook': {
                return {
                    webhook: {
                        id: action.id || null,
                        update: convertWebhookAction(action),
                    },
                }
            }
        }
    })
}
