import { action } from '@storybook/addon-actions'
import { number, select, text } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import React from 'react'
import { of } from 'rxjs'
import { NotificationType as NotificationTypeType } from 'sourcegraph'

import { NotificationType } from '@sourcegraph/extension-api-classes'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { NotificationItem } from './NotificationItem'

const notificationClassNames = {
    [NotificationType.Log]: 'alert alert-secondary',
    [NotificationType.Success]: 'alert alert-success',
    [NotificationType.Info]: 'alert alert-info',
    [NotificationType.Warning]: 'alert alert-warning',
    [NotificationType.Error]: 'alert alert-danger',
}

const onDismiss = action('onDismiss')

const decorator: DecoratorFn = story => (
    <>
        <style>{webStyles}</style>
        <div style={{ maxWidth: '20rem', margin: '2rem' }}>{story()}</div>
    </>
)
const config: Meta = {
    title: 'shared/NotificationItem',
    decorators: [decorator],
}
export default config

export const WithoutProgress: Story = () => {
    const message = text('Message', 'My *custom* message')
    const type = select<NotificationTypeType>(
        'Type',
        NotificationType as Record<keyof typeof NotificationType, NotificationTypeType>,
        NotificationType.Error
    )
    const source = text('Source', 'some source')
    return (
        <NotificationItem
            notification={{ message, type, source }}
            notificationClassNames={notificationClassNames}
            onDismiss={onDismiss}
        />
    )
}

export const WithProgress: Story = () => {
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
                progress: of({
                    message: progressMessage,
                    percentage: progressPercentage,
                }),
            }}
            notificationClassNames={notificationClassNames}
            onDismiss={onDismiss}
        />
    )
}

WithProgress.storyName = 'With progress'
