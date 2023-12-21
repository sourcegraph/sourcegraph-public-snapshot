import * as React from 'react'

import classNames from 'classnames'

import { asError } from '@sourcegraph/common'
import { Button, type ButtonProps, Heading, Tooltip, ErrorAlert, type HeadingElement } from '@sourcegraph/wildcard'

import styles from './ActionContainer.module.scss'

export const BaseActionContainer: React.FunctionComponent<
    React.PropsWithChildren<{
        title: React.ReactNode
        titleAs?: HeadingElement
        titleStyleAs?: HeadingElement
        description: React.ReactNode
        action?: React.ReactNode
        details?: React.ReactNode
        className?: string
    }>
> = ({ title, description, action, details, className, titleAs = 'h4', titleStyleAs = titleAs }) => (
    <div className={classNames(styles.actionContainer, className)}>
        <div className={styles.row}>
            <div className={styles.content}>
                <Heading as={titleAs} styleAs={titleStyleAs} className={styles.title}>
                    {title}
                </Heading>
                {description}
            </div>
            {action && <div className={styles.btnContainer}>{action}</div>}
        </div>
        {details && <div className={styles.row}>{details}</div>}
    </div>
)

interface Props {
    title: React.ReactNode
    titleAs?: HeadingElement
    titleStyleAs?: HeadingElement
    description: React.ReactNode
    buttonVariant?: ButtonProps['variant']
    buttonLabel: React.ReactNode
    buttonSubtitle?: string
    buttonDisabled?: boolean
    info?: React.ReactNode
    className?: string

    /** The message to briefly display below the button when the action is successful. */
    flashText?: string

    run: () => Promise<void>
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
                titleAs={this.props.titleAs}
                titleStyleAs={this.props.titleStyleAs}
                description={this.props.description}
                className={this.props.className}
                action={
                    <>
                        <Tooltip content={this.props.buttonSubtitle}>
                            <Button
                                className={styles.btn}
                                variant={this.props.buttonVariant || 'primary'}
                                onClick={this.onClick}
                                disabled={this.props.buttonDisabled || this.state.loading}
                            >
                                {this.props.buttonLabel}
                            </Button>
                        </Tooltip>
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
