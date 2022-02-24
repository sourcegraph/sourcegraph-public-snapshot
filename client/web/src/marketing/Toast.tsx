import classNames from 'classnames'
import CloseIcon from 'mdi-react/CloseIcon'
import * as React from 'react'

import { Button, CardTitle, CardBody, Card, Icon } from '@sourcegraph/wildcard'

import styles from './Toast.module.scss'

interface ToastProps {
    title: React.ReactNode
    subtitle?: React.ReactNode
    cta?: JSX.Element
    footer?: JSX.Element
    onDismiss: () => void
}

export const Toast: React.FunctionComponent<ToastProps> = props => (
    <Card className={styles.toast}>
        <CardBody>
            <CardTitle as="header" className={classNames(styles.header)}>
                <h2 className="mb-0">{props.title}</h2>
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
