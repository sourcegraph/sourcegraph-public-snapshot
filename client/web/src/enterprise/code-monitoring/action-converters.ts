import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import {
    IMonitorEmail,
    IMonitorEmailInput,
    IMonitorSlackWebhook,
    IMonitorSlackWebhookInput,
    IMonitorWebhook,
    IMonitorWebhookInput,
} from '@sourcegraph/shared/src/schema'

import {
    CodeMonitorFields,
    MonitorActionInput,
    MonitorEditActionInput,
    MonitorEmailPriority,
} from '../../graphql-operations'

import { MonitorAction } from './components/FormActionArea'

function isActionSupported(action: MonitorAction): action is IMonitorEmail | IMonitorSlackWebhook | IMonitorWebhook {
    // We currently support email, Slack webhook, and generic webhook actions
    return (
        action.__typename === 'MonitorEmail' ||
        action.__typename === 'MonitorSlackWebhook' ||
        action.__typename === 'MonitorWebhook'
    )
}

function convertEmailAction(action: IMonitorEmail, authenticatedUserId: AuthenticatedUser['id']): IMonitorEmailInput {
    return {
        enabled: action.enabled,
        includeResults: action.includeResults,
        priority: MonitorEmailPriority.NORMAL,
        recipients: [authenticatedUserId],
        header: '',
    }
}

function convertSlackWebhookAction(action: IMonitorSlackWebhook): IMonitorSlackWebhookInput {
    return {
        enabled: action.enabled,
        includeResults: action.includeResults,
        url: action.url,
    }
}

function convertWebhookAction(action: IMonitorWebhook): IMonitorWebhookInput {
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
    return actions.filter(isActionSupported).map(action => {
        switch (action.__typename) {
            case 'MonitorEmail':
                return {
                    email: convertEmailAction(action, authenticatedUserId),
                }
            case 'MonitorSlackWebhook':
                return {
                    slackWebhook: convertSlackWebhookAction(action),
                }
            case 'MonitorWebhook':
                return {
                    webhook: convertWebhookAction(action),
                }
        }
    })
}

export function convertActionsForUpdate(
    actions: CodeMonitorFields['actions']['nodes'],
    authenticatedUserId: AuthenticatedUser['id']
): MonitorEditActionInput[] {
    return actions.filter(isActionSupported).map(action => {
        // Convert empty IDs to null so action is created
        switch (action.__typename) {
            case 'MonitorEmail':
                return {
                    email: {
                        id: action.id || null,
                        update: convertEmailAction(action, authenticatedUserId),
                    },
                }
            case 'MonitorSlackWebhook':
                return {
                    slackWebhook: {
                        id: action.id || null,
                        update: convertSlackWebhookAction(action),
                    },
                }
            case 'MonitorWebhook':
                return {
                    webhook: {
                        id: action.id || null,
                        update: convertWebhookAction(action),
                    },
                }
        }
    })
}
