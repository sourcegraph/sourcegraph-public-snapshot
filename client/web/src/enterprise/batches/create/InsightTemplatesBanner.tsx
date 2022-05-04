import React from 'react'

import classNames from 'classnames'

import { Card, CardBody, H4 } from '@sourcegraph/wildcard'

import { CodeInsightsBatchesIcon } from './CodeInsightsBatchesIcon'

import styles from './InsightTemplatesBanner.module.scss'

export const InsightTemplatesBanner: React.FunctionComponent<{ insightTitle: string }> = ({ insightTitle }) => (
    <Card className={classNames('mb-5', styles.banner)}>
        <CardBody>
            <div className="d-flex justify-content-between align-items-center">
                <CodeInsightsBatchesIcon className="mr-4" />
                <div className="flex-grow-1">
                    <H4>You are creating a batch change from a code insight</H4>
                    <p className="mb-0">
                        Let Sourcegraph help you with <strong>{insightTitle}</strong> by preparing a relevant{' '}
                        <strong>batch change</strong>.
                    </p>
                </div>
            </div>
        </CardBody>
    </Card>
)
