import * as React from 'react'

import { mdiContentCopy, mdiEye } from '@mdi/js'
import classNames from 'classnames'
import copy from 'copy-to-clipboard'

import { Button, Icon, Input } from '@sourcegraph/wildcard'

import styles from './CopyableText.module.scss'

interface Props {
    /** The text to present and to copy. */
    text: string

    children?: (props: { isRedacted: boolean }) => React.ReactNode

    /** An optional class name. */
    className?: string

    /** Whether the input should take up all horizontal space (flex:1) */
    flex?: boolean

    /** The size of the input element. */
    size?: number

    /** Whether the text to be copied is a password. */
    password?: boolean

    /** The label used for screen readers */
    label?: string

    /** Callback for when the content is copied  */
    onCopy?: () => void

    /** Whether the text is a secret, i.e. supports being shown/concealed with a click of a button */
    secret?: boolean
}

interface State {
    /** Whether the text was just copied. */
    copied: boolean

    /** Whether the secret is shown. */
    secretShown: boolean
}

/**
 * A component that displays a single line of text and a copy-to-clipboard button. There are other
 * niceties, such as triple-clicking selects only the text and not other adjacent components' text
 * labels.
 */
export class CopyableText extends React.PureComponent<Props, State> {
    public state: State = {
        copied: false,
        secretShown: false,
    }

    public render(): JSX.Element | null {
        return (
            <>
                <div className={classNames('form-inline', this.props.className)}>
                    <div className={classNames('input-group', { 'flex-1': this.props.flex })}>
                        <Input
                            type={this.resolveInputType()}
                            inputClassName={styles.input}
                            aria-label={this.props.label}
                            value={this.props.text}
                            size={this.props.size}
                            readOnly={true}
                            onClick={this.onClickInput}
                        />
                        <div className="input-group-append flex-shrink-0">
                            <Button
                                onClick={this.onClickButton}
                                disabled={this.state.copied}
                                variant="secondary"
                                aria-label="Copy"
                            >
                                <Icon aria-hidden={true} svgPath={mdiContentCopy} />{' '}
                                {this.props.secret ? '' : this.state.copied ? 'Copied' : 'Copy'}
                            </Button>
                        </div>
                        {this.props.secret && (
                            <div className="input-group-append flex-shrink-0 ml-1">
                                <Button
                                    onClick={this.onClickSecretButton}
                                    variant="secondary"
                                    aria-label={
                                        this.state.secretShown ? 'Reveal secret value as text' : 'Hide secret value'
                                    }
                                >
                                    <Icon aria-hidden={true} svgPath={mdiEye} />
                                </Button>
                            </div>
                        )}
                    </div>
                </div>

                {this.props.children?.({ isRedacted: Boolean(this.props.secret && !this.state.secretShown) })}
            </>
        )
    }

    private onClickInput: React.MouseEventHandler<HTMLInputElement> = event => {
        event.currentTarget.focus()
        event.currentTarget.setSelectionRange(0, this.props.text.length)
        this.copyToClipboard()
    }

    private onClickButton = (): void => this.copyToClipboard()

    private copyToClipboard(): void {
        copy(this.props.text)
        this.setState({ copied: true })

        setTimeout(() => this.setState({ copied: false }), 1000)

        if (typeof this.props.onCopy === 'function') {
            this.props.onCopy()
        }
    }

    private onClickSecretButton = (): void =>
        this.setState(state => ({
            secretShown: !state.secretShown,
        }))

    private resolveInputType(): string {
        if (this.props.password) {
            return 'password'
        }
        if (this.props.secret) {
            return this.state.secretShown ? 'text' : 'password'
        }
        return 'text'
    }
}
