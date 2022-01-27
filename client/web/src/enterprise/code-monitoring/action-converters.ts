import { AuthenticatedUser } from '@sourcegraph/shared/src/auth'
import {
    IMonitorEmail,
    IMonitorEmailInput,
    IMonitorSlackWebhook,
    IMonitorSlackWebhookInput,
} from '@sourcegraph/shared/src/schema'

import {
    CodeMonitorFields,
    MonitorActionInput,
    MonitorEditActionInput,
    MonitorEmailPriority,
} from '../../graphql-operations'

import { MonitorAction } from './components/FormActionArea'

function isActionSupported(action: MonitorAction): action is IMonitorEmail | IMonitorSlackWebhook {
    // We currently only support email and Slack webhook actions
    return action.__typename === 'MonitorEmail' || action.__typename === 'MonitorSlackWebhook'
}

function convertEmailAction(action: IMonitorEmail, authenticatedUserId: AuthenticatedUser['id']): IMonitorEmailInput {
    return {
        enabled: action.enabled,
        priority: MonitorEmailPriority.NORMAL,
        recipients: [authenticatedUserId],
        header: '',
    }
}

function convertSlackWebhookAction(action: IMonitorSlackWebhook): IMonitorSlackWebhookInput {
    return {
        enabled: action.enabled,
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
        }
    })
}
