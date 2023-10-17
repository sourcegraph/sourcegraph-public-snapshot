import React, { useCallback, useState } from 'react'

import { mdiChevronDown, mdiChevronUp } from '@mdi/js'
import classNames from 'classnames'
import { format } from 'date-fns'
import type { Optional } from 'utility-types'

import { CodeSnippet } from '@sourcegraph/branded/src/components/CodeSnippet'
import { Button, Tab, TabList, TabPanel, TabPanels, Tabs, Icon } from '@sourcegraph/wildcard'

import type { WebhookLogFields } from '../../graphql-operations'

import { MessagePanel } from './MessagePanel'
import { StatusCode } from './StatusCode'

import styles from './WebhookLogNode.module.scss'

export interface Props {
    node: Optional<WebhookLogFields, 'response'> & { error?: string; eventType?: string }
    doNotShowExternalService?: boolean

    // For storybook purposes only:
    initiallyExpanded?: boolean
    initialTabIndex?: number
}

export const WebhookLogNode: React.FunctionComponent<React.PropsWithChildren<Props>> = ({
    doNotShowExternalService = false,
    initiallyExpanded,
    initialTabIndex,
    node: { error, eventType, externalService, receivedAt, request, response, statusCode },
}) => {
    const [isExpanded, setIsExpanded] = useState(initiallyExpanded === true)
    const toggleExpanded = useCallback(() => setIsExpanded(!isExpanded), [isExpanded])

    return (
        <>
            <span className={styles.separator} />
            <span className={styles.detailsButton}>
                <Button
                    variant="icon"
                    aria-label={isExpanded ? 'Collapse section' : 'Expand section'}
                    onClick={toggleExpanded}
                >
                    <Icon aria-hidden={true} svgPath={isExpanded ? mdiChevronUp : mdiChevronDown} />
                </Button>
            </span>
            <span className={styles.statusCode}>
                <StatusCode code={statusCode} />
            </span>
            <span>
                {!doNotShowExternalService ? (
                    externalService ? (
                        externalService.displayName
                    ) : (
                        <span className="text-danger">Unmatched</span>
                    )
                ) : (
                    eventType ?? undefined
                )}
            </span>
            <span className={styles.receivedAt}>{format(Date.parse(receivedAt), 'Ppp')}</span>
            <span className={styles.smDetailsButton}>
                <Button onClick={toggleExpanded} outline={true} variant="secondary">
                    {isExpanded ? (
                        <Icon aria-hidden={true} svgPath={mdiChevronUp} />
                    ) : (
                        <Icon aria-hidden={true} svgPath={mdiChevronDown} />
                    )}{' '}
                    {isExpanded ? 'Hide' : 'Show'} details
                </Button>
            </span>
            {isExpanded && (
                <div className={classNames('px-4', 'pt-3', 'pb-2', styles.expanded)}>
                    <Tabs index={initialTabIndex} size="small">
                        <TabList>
                            <Tab>Request</Tab>
                            <Tab>{response ? 'Response' : 'Error'}</Tab>
                        </TabList>
                        <TabPanels>
                            <TabPanel>
                                <MessagePanel
                                    className={styles.messagePanelContainer}
                                    message={request}
                                    requestOrStatusCode={request}
                                />
                            </TabPanel>
                            <TabPanel>
                                {response ? (
                                    <MessagePanel
                                        className={styles.messagePanelContainer}
                                        message={response}
                                        requestOrStatusCode={statusCode}
                                    />
                                ) : (
                                    <CodeSnippet language="nohighlight" code={error ?? ''} />
                                )}
                            </TabPanel>
                        </TabPanels>
                    </Tabs>
                </div>
            )}
        </>
    )
}
