import { action } from '@storybook/addon-actions'
import { DecoratorFn, Meta, Story } from '@storybook/react'
import { of } from 'rxjs'
import { NotificationType as NotificationTypeType } from 'sourcegraph'

import { BrandedStory } from '@sourcegraph/branded/src/components/BrandedStory'
import { NotificationType } from '@sourcegraph/extension-api-classes'

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
    <BrandedStory>{() => <div style={{ maxWidth: '20rem', margin: '2rem' }}>{story()}</div>}</BrandedStory>
)

const config: Meta = {
    title: 'shared/NotificationItem',
    decorators: [decorator],
    argTypes: {
        message: {
            name: 'Message',
            control: { type: 'text' },
            defaultValue: 'My *custom* message',
        },
        type: {
            name: 'type',
            control: {
                type: 'select',
                options: NotificationType as Record<keyof typeof NotificationType, NotificationTypeType>,
            },
        },
        source: {
            name: 'Source',
            control: { type: 'text' },
            defaultValue: 'some source',
        },
    },
}
export default config

export const WithoutProgress: Story = args => {
    const message = args.message
    const type = args.type
    const source = args.source
    return (
        <NotificationItem
            notification={{ message, type, source }}
            notificationItemStyleProps={{ notificationItemClassNames: notificationClassNames }}
            onDismiss={onDismiss}
        />
    )
}
WithoutProgress.argTypes = {
    type: {
        defaultValue: NotificationType.Error,
    },
}
export const WithProgress: Story = args => {
    const message = args.message
    const type = args.type
    const source = args.source
    const progressMessage = args.progressMessage
    const progressPercentage = args.progressPercentage
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
WithProgress.argTypes = {
    progressMessage: {
        name: 'Progress message',
        control: { type: 'text' },
        defaultValue: 'My *custom* progress message',
    },
    progressPercentage: {
        name: 'Progress % (0-100)',
        control: { type: 'number', min: 0, max: 100 },
        defaultValue: 50,
    },
    type: {
        defaultValue: NotificationType.Info,
    },
}

WithProgress.storyName = 'With progress'
WithProgress.parameters = {
    chromatic: {
        enableDarkMode: true,
        disableSnapshot: false,
    },
}
