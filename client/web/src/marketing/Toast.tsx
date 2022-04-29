import * as React from 'react'

import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'

import { Button, CardTitle, CardBody, Card, Icon } from '@sourcegraph/wildcard'

import styles from './Toast.module.scss'

interface ToastProps {
    title?: React.ReactNode
    subtitle?: React.ReactNode
    cta?: JSX.Element
    footer?: JSX.Element
    onDismiss: () => void
    className?: string
}

export const Toast: React.FunctionComponent<ToastProps> = props => (
    <Card className={classNames(styles.toast, props.className)}>
        <CardBody className={styles.toastBody}>
            <div className={classNames('d-flex justify-content-end', styles.closeButtonWrap)}>
                <Button onClick={props.onDismiss} variant="icon" className="test-close-toast" aria-label="Close">
                    <Icon as={CloseIcon} />
                </Button>
            </div>
            <CardTitle as="header" className={styles.header}>
                {props.title && <h2 className="mb-0">{props.title}</h2>}
            </CardTitle>
            {props.subtitle}
            {props.cta && <div className={styles.contentsCta}>{props.cta}</div>}
        </CardBody>
        {props.footer && <div className={classNames(styles.footer)}>{props.footer}</div>}
    </Card>
)
