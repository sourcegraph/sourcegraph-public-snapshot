import * as React from 'react'

import classNames from 'classnames'

import { LinkOrSpan } from '@sourcegraph/shared/src/components/LinkOrSpan'
import { CardText, CardTitle, CardBody, Card, H4 } from '@sourcegraph/wildcard'

import styles from './SingleValueCard.module.scss'

/**
 * A card that displays a large single value.
 */
export const SingleValueCard: React.FunctionComponent<
    React.PropsWithChildren<{
        title: string
        subTitle?: string
        value: string | number
        link?: string
        className?: string
        valueClassName?: string
        valueTooltip?: string
        subText?: string
    }>
> = ({ title, value, subTitle, link, className, valueClassName, valueTooltip, subText }) => (
    <Card className={classNames(styles.singleValueCard, className)}>
        <CardBody className="text-center">
            <CardTitle as={H4} className="mb-0">
                {title}
            </CardTitle>
            <CardText as="small">{subTitle || ''}</CardText>
            <CardText
                data-tooltip={valueTooltip}
                className={classNames(classNames('font-weight-bold text-nowrap', styles.value), valueClassName)}
            >
                <LinkOrSpan to={link}>{value}</LinkOrSpan>
            </CardText>
            {subText && <small>{subText}</small>}
        </CardBody>
    </Card>
)
