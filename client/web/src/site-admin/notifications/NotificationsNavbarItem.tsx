import React, { useMemo } from 'react'

import { mdiBellAlert, mdiBellBadge, mdiBellCheck, mdiCheckCircle, mdiInformation } from '@mdi/js'
import classNames from 'classnames'
import { groupBy, startCase } from 'lodash'

import { renderMarkdown } from '@sourcegraph/common'
import {
    Icon,
    Badge,
    Popover,
    PopoverTrigger,
    PopoverContent,
    Text,
    Button,
    PopoverTail,
    Alert,
    Code,
} from '@sourcegraph/wildcard'

import { AlertType } from '../../graphql-operations'

import { useNotifications } from './hooks/useNotifications'

import styles from './NotificationsNavbarItem.module.scss'

const NotificationsGroup: React.FC<{ name: string; items: { message: string; type: AlertType }[] }> = ({
    name,
    items,
}) => {
    const hasError = items.some(item => item.type === AlertType.ERROR)
    const hasWarning = !hasError && items.some(item => item.type === AlertType.WARNING)
    const hasInfo = !hasWarning && items.some(item => item.type === AlertType.INFO)
    return (
        <div>
            <div className="mb-2">
                <Icon
                    svgPath={hasError || hasWarning || hasInfo ? mdiInformation : mdiCheckCircle}
                    aria-hidden={true}
                    className={classNames(styles.icon, {
                        [styles.danger]: hasError,
                        [styles.warning]: hasWarning,
                        [styles.info]: hasInfo,
                    })}
                />
                {name}
            </div>
            <div className="pl-2">
                {items.map(({ message, type }) => (
                    <Text
                        key={message}
                        className={classNames(styles.message, {
                            [styles.warning]: type === AlertType.WARNING,
                            [styles.danger]: type === AlertType.ERROR,
                            [styles.info]: type === AlertType.INFO,
                        })}
                        dangerouslySetInnerHTML={{ __html: renderMarkdown(message) }}
                    />
                ))}
            </div>
        </div>
    )
}

export const NotificationsNavbarItem: React.FC = () => {
    const { data, loading, error } = useNotifications()
    const notConfiguredCount = data.length
    const groups = useMemo(() => groupBy(data, 'group'), [data])

    if (loading) {
        return null
    }

    const hasError = data.find(item => item.type === AlertType.ERROR)
    const hasWarning = data.find(item => item.type === AlertType.WARNING)
    const hasInfo = data.find(item => item.type === AlertType.INFO)

    return (
        <Popover>
            <PopoverTrigger as={Button}>
                <div className="d-flex align-items-center">
                    <Icon
                        svgPath={(hasError ? mdiBellAlert : hasWarning || hasInfo) ? mdiBellBadge : mdiBellCheck}
                        aria-label="Notifications"
                        size="md"
                    />
                    {notConfiguredCount > 0 && (
                        <sup>
                            <Badge
                                variant={hasError ? 'danger' : hasWarning ? 'warning' : hasInfo ? 'info' : 'secondary'}
                                pill={true}
                                small={true}
                            >
                                {notConfiguredCount}
                            </Badge>
                        </sup>
                    )}
                </div>
            </PopoverTrigger>
            <PopoverContent className={classNames('p-3', styles.popoverContent)}>
                {error && (
                    <Alert variant="danger">
                        Error happened while loading notifications: <br />
                        <Code>{JSON.stringify(error, null, 2)}</Code>
                    </Alert>
                )}
                {Object.entries(groups).map(([group, items]) => (
                    <NotificationsGroup name={startCase(group.toLowerCase())} items={items} key={group} />
                ))}
            </PopoverContent>
            <PopoverTail />
        </Popover>
    )
}
