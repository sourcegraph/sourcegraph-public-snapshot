import React from 'react'

import { Card, CardBody, H4, Text } from '@sourcegraph/wildcard'

import { CodeInsightsBatchesIcon } from './CodeInsightsBatchesIcon'

import styles from './TemplateBanner.module.scss'

interface TemplatesBannerProps {
    heading: string | React.ReactNode
    description: string | React.ReactNode
}

export const TemplateBanner: React.FunctionComponent<TemplatesBannerProps> = ({ heading, description }) => (
    <Card className={styles.banner}>
        <CardBody>
            <div className="d-flex justify-content-between align-items-center">
                <CodeInsightsBatchesIcon className={styles.icon} />
                <div className="flex-grow-1">
                    <H4 className="mb-1">{heading}</H4>
                    <Text className={styles.bannerDescription}>{description}</Text>
                </div>
            </div>
        </CardBody>
    </Card>
)
