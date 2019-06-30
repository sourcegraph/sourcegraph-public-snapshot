import * as sourcegraph from 'sourcegraph'
import { Unsubscribable, Subscription } from 'rxjs'

const STATUSES: (sourcegraph.Status & { name: string })[] = [
    {
        name: 'code-churn',
        title: 'Code churn',
        state: {
            completion: sourcegraph.StatusCompletion.Completed,
            result: sourcegraph.StatusResult.ActionRequired,
            message: 'High code churn detected',
        },
        notifications: [
            { title: 'my notif1', type: sourcegraph.NotificationType.Info },
            { title: 'my notif2', type: sourcegraph.NotificationType.Error },
        ],
    },
    {
        name: 'eslint',
        title: 'ESLint',
        state: {
            completion: sourcegraph.StatusCompletion.Completed,
            result: sourcegraph.StatusResult.Success,
            message: 'Compliant with up-to-date ESLint rules',
        },
        notifications: [
            { title: 'my notif1', type: sourcegraph.NotificationType.Info },
            { title: 'my notif2', type: sourcegraph.NotificationType.Error },
        ],
    },
    {
        name: 'npm-dependency-security',
        title: 'npm dependency security',
        state: {
            completion: sourcegraph.StatusCompletion.InProgress,
            message: 'Scanning npm dependencies...',
        },
        notifications: [
            { title: 'my notif1', type: sourcegraph.NotificationType.Info },
            { title: 'my notif2', type: sourcegraph.NotificationType.Error },
        ],
    },
]

export function registerSampleStatusProviders(): Unsubscribable {
    const subscriptions = new Subscription()
    for (const status of STATUSES) {
        subscriptions.add(
            sourcegraph.status.registerStatusProvider(status.name, {
                provideStatus: () => status,
            })
        )
    }
    return subscriptions
}
