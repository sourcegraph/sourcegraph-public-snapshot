import * as React from 'react'

import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button, CardTitle, CardBody, Card, Icon, Typography } from '@sourcegraph/wildcard'

import styles from './Toast.module.scss'

interface ToastProps {
    title: React.ReactNode
    subtitle?: React.ReactNode
    cta?: JSX.Element
    footer?: JSX.Element
    onDismiss: () => void
}

export const Toast: React.FunctionComponent<React.PropsWithChildren<ToastProps>> = props => (
    <Card className={styles.toast}>
        <CardBody>
            <CardTitle as="header" className={classNames(styles.header)}>
                <Typography.H2 className="mb-0">{props.title}</Typography.H2>
                <Button
                    onClick={props.onDismiss}
                    variant="icon"
                    className={classNames('test-close-toast', styles.closeButton)}
                    aria-label="Close"
                >
                    <Icon as={CloseIcon} />
                </Button>
            </CardTitle>
            {props.subtitle}
            {props.cta && <div className={styles.contentsCta}>{props.cta}</div>}
        </CardBody>
        {props.footer && <div className={classNames(styles.footer)}>{props.footer}</div>}
    </Card>
)
