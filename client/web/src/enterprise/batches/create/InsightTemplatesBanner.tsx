import React from 'react'

import classNames from 'classnames'

import { Card, CardBody, Typography } from '@sourcegraph/wildcard'

import { CodeInsightsBatchesIcon } from './CodeInsightsBatchesIcon'

import styles from './InsightTemplatesBanner.module.scss'

interface InsightTemplatesBannerProps {
    insightTitle: string
    type: 'create' | 'edit'
    className?: string
}

export const InsightTemplatesBanner: React.FunctionComponent<React.PropsWithChildren<InsightTemplatesBannerProps>> = ({
    insightTitle,
    className,
    type,
}) => {
    const [heading, paragraph]: [React.ReactNode, React.ReactNode] =
        type === 'create'
            ? [
                  'You are creating a batch change from a code insight',
                  <>
                      Let Sourcegraph help you with <strong>{insightTitle}</strong> by preparing a relevant{' '}
                      <strong>batch change</strong>.
                  </>,
              ]
            : [
                  `Start from template for the ${insightTitle}`,
                  `Sourcegraph pre-selected a batch spec for the batch change started from ${insightTitle}.`,
              ]

    return (
        <Card className={classNames(className, styles.banner)}>
            <CardBody>
                <div className="d-flex justify-content-between align-items-center">
                    <CodeInsightsBatchesIcon className="mr-4" />
                    <div className="flex-grow-1">
                        <Typography.H4>{heading}</Typography.H4>
                        <p className="mb-0">{paragraph}</p>
                    </div>
                </div>
            </CardBody>
        </Card>
    )
}
