import React from 'react'

import classNames from 'classnames'

import { Card, CardBody, H4, Text } from '@sourcegraph/wildcard'

import { BatchChangeTemplateIcon } from './BatchChangeTemplateIcon'

import styles from './TemplateBanner.module.scss'

interface TemplatesBannerProps {
    heading: string | React.ReactNode
    description: string | React.ReactNode
    className?: string
}

export const TemplateBanner: React.FunctionComponent<TemplatesBannerProps> = ({ heading, description, className }) => (
    <Card className={classNames(className, styles.banner)}>
        <CardBody>
            <div className="d-flex justify-content-between align-items-center">
                <BatchChangeTemplateIcon className={styles.icon} />
                <div className="flex-grow-1">
                    <H4 className="mb-1">{heading}</H4>
                    <Text className={styles.bannerDescription}>{description}</Text>
                </div>
            </div>
        </CardBody>
    </Card>
)
