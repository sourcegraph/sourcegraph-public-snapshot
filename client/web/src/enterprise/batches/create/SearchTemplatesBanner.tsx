import React from 'react'

import { Card, CardBody, H4, Text } from '@sourcegraph/wildcard'

import { CodeInsightsBatchesIcon } from './CodeInsightsBatchesIcon'

import styles from './SearchTemplatesBanner.module.scss'

export const SearchTemplatesBanner: React.FunctionComponent = () => (
    <Card className={styles.banner}>
        <CardBody className="d-flex justify-content-between align-items-center">
            <CodeInsightsBatchesIcon className={styles.icon} />
            <div className="flex-grow-1 justify-content-between align-items-center">
                <H4 className="mb-1">You are creating a Batch Change from a Code Search</H4>
                <Text className={styles.bannerDescription}>
                    Let Sourcegraph help you refactor your code by preparing a Batch Change from your search query
                </Text>
            </div>
        </CardBody>
    </Card>
)
