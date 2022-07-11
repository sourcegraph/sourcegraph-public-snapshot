import React from 'react'

import classNames from 'classnames'

import { Card, CardBody, H4, Text } from '@sourcegraph/wildcard'

import { BatchChangeTemplateIcon } from './BatchChangeTemplateIcon'

import styles from './TemplateBanner.module.scss'

interface TemplatesBannerProps {
    /**
     * Title of the template the batch change is created from.
     */
    heading: string | React.ReactNode

    /**
     * Description of the template used to create a batch change.
     */
    description: string | React.ReactNode

    /**
     * CSS class to be applied to banner wrapper. This classname is passed into the
     * outermost card component.
     */
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
