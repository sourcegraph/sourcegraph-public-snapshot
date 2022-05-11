import { action } from '@storybook/addon-actions'
import { number, select, text } from '@storybook/addon-knobs'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import { of } from 'rxjs'
import { NotificationType as NotificationTypeType } from 'sourcegraph'

import { NotificationType } from '@sourcegraph/extension-api-classes'
import webStyles from '@sourcegraph/web/src/SourcegraphWebApp.scss'

import { NotificationItem } from './NotificationItem'

const notificationClassNames = {
    [NotificationType.Log]: 'bg-secondary',
    [NotificationType.Success]: 'bg-success',
    [NotificationType.Info]: 'bg-info',
    [NotificationType.Warning]: 'bg-warning',
    [NotificationType.Error]: 'bg-danger',
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
            notificationItemStyleProps={{ notificationItemClassNames: notificationClassNames }}
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
            notificationItemStyleProps={{ notificationItemClassNames: notificationClassNames }}
            onDismiss={onDismiss}
        />
    )
}

WithProgress.storyName = 'With progress'
WithProgress.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
