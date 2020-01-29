import { NotificationType } from '@sourcegraph/extension-api-classes'
import { action } from '@storybook/addon-actions'
import { number, select, text } from '@storybook/addon-knobs'
import { storiesOf } from '@storybook/react'
import React from 'react'
import { interval } from 'rxjs'
import { map, startWith } from 'rxjs/operators'
import { NotificationType as NotificationTypeType } from 'sourcegraph'
import { NotificationItem } from './NotificationItem'

import './NotificationItem.scss'

const notificationClassNames = {
    [NotificationType.Log]: 'alert alert-secondary',
    [NotificationType.Success]: 'alert alert-success',
    [NotificationType.Info]: 'alert alert-info',
    [NotificationType.Warning]: 'alert alert-warning',
    [NotificationType.Error]: 'alert alert-error',
}

const onDismiss = action('onDismiss')

const { add } = storiesOf('NotificationItem', module).addDecorator(story => (
    // tslint:disable-next-line: jsx-ban-props
    <div style={{ maxWidth: '20rem', margin: '2rem' }}>{story()}</div>
))

for (const [name, type] of Object.entries(NotificationType)) {
    // TS enums are reverse-indexed, so filter the number keys out
    if (!isNaN(parseInt(name, 10))) {
        continue
    }

    add(name, () => (
        <NotificationItem
            notification={{
                message: 'Formatted *message*',
                type: type as NotificationTypeType,
            }}
            notificationClassNames={notificationClassNames}
            onDismiss={onDismiss}
        />
    ))

    add(`${name} - Progress`, () => (
        <NotificationItem
            notification={{
                type: type as NotificationTypeType,
                progress: interval(100).pipe(
                    startWith(0),
                    map(i => ({
                        message: 'Formatted progress *message*',
                        percentage: i % 25 < 15 ? (i + 15) % 101 : undefined,
                    }))
                ),
            }}
            notificationClassNames={notificationClassNames}
            onDismiss={onDismiss}
        />
    ))
}

add('⚙', () => {
    const message = text('Message', 'My *custom* message')
    const type = select<NotificationTypeType>(
        'Type',
        NotificationType as Record<keyof typeof NotificationType, NotificationTypeType>,
        NotificationType.Info
    )
    const source = text('Source', 'some source')
    const progressMessage = text('Progress message', 'My *custom* progress message')
    const progressPercentage = number('Progress % (0-100)', 50)
    return (
        <NotificationItem
            notification={{
                message,
                type,
                source,
                progress: interval(1000).pipe(
                    startWith(0),
                    map(i => ({
                        message: progressMessage,
                        percentage: progressPercentage,
                    }))
                ),
            }}
            notificationClassNames={notificationClassNames}
            onDismiss={onDismiss}
        />
    )
})
