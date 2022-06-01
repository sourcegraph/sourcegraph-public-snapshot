import * as React from 'react'

import classNames from 'classnames'
import * as H from 'history'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { asError } from '@sourcegraph/common'
import { Button, Typography } from '@sourcegraph/wildcard'

import styles from './ActionContainer.module.scss'

export const BaseActionContainer: React.FunctionComponent<
    React.PropsWithChildren<{
        title: React.ReactFragment
        description: React.ReactFragment
        action: React.ReactFragment
        details?: React.ReactFragment
        className?: string
    }>
> = ({ title, description, action, details, className }) => (
    <div className={classNames(styles.actionContainer, className)}>
        <div className={styles.row}>
            <div>
                <Typography.H4 className={styles.title}>{title}</Typography.H4>
                {description}
            </div>
            <div className={styles.btnContainer}>{action}</div>
        </div>
        {details && <div className={styles.row}>{details}</div>}
    </div>
)

interface Props {
    title: React.ReactFragment
    description: React.ReactFragment
    buttonClassName?: string
    buttonLabel: React.ReactFragment
    buttonSubtitle?: string
    buttonDisabled?: boolean
    info?: React.ReactNode
    className?: string

    /** The message to briefly display below the button when the action is successful. */
    flashText?: string

    run: () => Promise<void>
    history: H.History
}

interface State {
    loading: boolean
    flash: boolean
    error?: string
}

/**
 * Displays an action button in a container with a title and description.
 */
export class ActionContainer extends React.PureComponent<Props, State> {
    public state: State = {
        loading: false,
        flash: false,
    }

    private timeoutHandle?: number

    public componentWillUnmount(): void {
        if (this.timeoutHandle) {
            window.clearTimeout(this.timeoutHandle)
        }
    }

    public render(): JSX.Element | null {
        return (
            <BaseActionContainer
                title={this.props.title}
                description={this.props.description}
                className={this.props.className}
                action={
                    <>
                        <Button
                            className={classNames(styles.btn, this.props.buttonClassName)}
                            variant={this.props.buttonClassName ? undefined : 'primary'}
                            onClick={this.onClick}
                            data-tooltip={this.props.buttonSubtitle}
                            disabled={this.props.buttonDisabled || this.state.loading}
                        >
                            {this.props.buttonLabel}
                        </Button>
                        {this.props.buttonSubtitle && (
                            <div className={styles.btnSubtitle}>
                                <small>{this.props.buttonSubtitle}</small>
                            </div>
                        )}
                        {!this.props.buttonSubtitle && this.props.flashText && (
                            <div className={classNames(styles.flash, this.state.flash && styles.flashVisible)}>
                                <small>{this.props.flashText}</small>
                            </div>
                        )}
                    </>
                }
                details={
                    <>
                        {this.state.error ? (
                            <ErrorAlert className="mb-0 mt-3" error={this.state.error} />
                        ) : (
                            this.props.info
                        )}
                    </>
                }
            />
        )
    }

    private onClick = (): void => {
        this.setState({
            error: undefined,
            loading: true,
        })

        this.props.run().then(
            () => {
                this.setState({ loading: false, flash: true })
                if (typeof this.timeoutHandle === 'number') {
                    window.clearTimeout(this.timeoutHandle)
                }
                this.timeoutHandle = window.setTimeout(() => this.setState({ flash: false }), 1000)
            },
            error => this.setState({ loading: false, error: asError(error).message })
        )
    }
}
