import * as React from 'react'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import DropdownItem from 'reactstrap/lib/DropdownItem'
import { from, Subject, Subscribable, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { ExtensionsProps } from '../context'
import { asError, ErrorLike, isErrorLike } from '../errors'
import {
    ConfigurationCascadeProps,
    ConfigurationSubject,
    ConfiguredSubjectOrError,
    Settings,
    subjectLabel,
} from '../settings'
import { ConfiguredExtension, isExtensionAdded, isExtensionEnabled } from './extension'

const LOADING: 'loading' = 'loading'

interface ExtensionConfigureDropdownItemState {
    /** The operation's status: null when done or not started, 'loading', or an error. */
    operationResultOrError: typeof LOADING | null | ErrorLike
}

/** An item in the {@link ExtensionConfigureButton} dropdown menu.  */
export class ExtensionConfigureDropdownItem<
    S extends ConfigurationSubject,
    C extends Settings
> extends React.PureComponent<
    {
        /** The extension that this button is for. */
        extension: ConfiguredExtension

        /** The configuration subject that this item modifies extension settings for. */
        subject: ConfiguredSubjectOrError<ConfigurationSubject, Settings>

        disabled?: boolean
        confirm?: () => boolean
        operation: (
            extension: ConfiguredExtension,
            subject: ConfiguredSubjectOrError<ConfigurationSubject, Settings>
        ) => Subscribable<void>
        onUpdate: () => void
        onComplete: () => void
    } & ExtensionsProps<S, C>,
    ExtensionConfigureDropdownItemState
> {
    public state: ExtensionConfigureDropdownItemState = { operationResultOrError: null }

    private clicks = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.clicks
                .pipe(
                    switchMap(() =>
                        from(this.props.operation(this.props.extension, this.props.subject)).pipe(
                            mapTo(null),
                            tap(() => this.props.onComplete()),
                            catchError(error => [asError(error) as ErrorLike]),
                            map(c => ({ operationResultOrError: c } as ExtensionConfigureDropdownItemState)),
                            tap(() => this.props.onUpdate()),
                            startWith<ExtensionConfigureDropdownItemState>({ operationResultOrError: LOADING })
                        )
                    )
                )
                .subscribe(stateUpdate => this.setState(stateUpdate), error => console.error(error))
        )
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        return (
            <DropdownItem
                className="dropdown-item d-flex justify-content-between align-items-center"
                disabled={this.props.disabled}
                onClick={this.onClick}
                toggle={false}
            >
                <span className="mr-5">{this.props.children}</span>
                <div>
                    {isErrorLike(this.state.operationResultOrError) && (
                        <small className="text-danger" title={this.state.operationResultOrError.message}>
                            <this.props.extensions.context.icons.Warning className="icon-inline" /> Error
                        </small>
                    )}
                </div>
            </DropdownItem>
        )
    }

    private onClick: React.MouseEventHandler<HTMLElement> = () => {
        if (!this.props.confirm || this.props.confirm()) {
            this.clicks.next()
        }
    }
}

interface Props<S extends ConfigurationSubject, C extends Settings>
    extends ConfigurationCascadeProps<S, C>,
        ExtensionsProps<S, C> {
    /** The extension that this dropdown is for. */
    extension: ConfiguredExtension

    /** The configuration subject that this dropdown modifies extension settings for. */
    subject: ConfiguredSubjectOrError<ConfigurationSubject, Settings>

    /** Class name applied to the button element. */
    buttonClassName?: string

    /* The button label. */
    children: React.ReactFragment

    /** Whether to show the caret on the dropdown toggle. */
    caret?: boolean

    /**
     * Called to confirm the primary action. If the callback returns false, the action is not
     * performed.
     */
    confirm?: () => boolean

    /** Called when the component performs an update that requires the parent component to refresh data. */
    onUpdate: () => void
}

interface State {
    dropdownOpen: boolean
}

/**
 * Displays a button with a dropdown menu for enabling/disabling the extension.
 *
 * For simplicity, the menu is only intended to expose the most common extension configuration actions for the
 * current user. For example, it does not expose actions to configure the extension for all users (in global
 * settings) or for an organization's members. To make those changes, the user needs to manually edit global or
 * organization settings.
 */
export class ExtensionConfigureButtonDropdown<
    S extends ConfigurationSubject,
    C extends Settings
> extends React.PureComponent<Props<S, C>, State> {
    public state: State = {
        dropdownOpen: false,
    }

    public render(): JSX.Element | null {
        // Configuration subjects other than this.props.subject for which the extension is added in settings.
        const otherSubjectsWithExtensionAdded =
            this.props.configurationCascade.subjects && !isErrorLike(this.props.configurationCascade.subjects)
                ? this.props.configurationCascade.subjects
                      .filter(a => a.subject.id !== this.props.subject.subject.id)
                      .filter(subject => isExtensionAdded(subject.settings, this.props.extension.id))
                : []

        return (
            <ButtonDropdown isOpen={this.state.dropdownOpen} toggle={this.toggle} group={this.props.caret !== false}>
                <DropdownToggle caret={this.props.caret !== false} color="" className={this.props.buttonClassName}>
                    {this.props.children}
                </DropdownToggle>
                <DropdownMenu>
                    <DropdownItem header={true}>{subjectLabel(this.props.subject.subject)} settings:</DropdownItem>
                    <ExtensionConfigureDropdownItem
                        extension={this.props.extension}
                        subject={this.props.subject}
                        operation={this.enableExtensionForSubject}
                        onUpdate={this.props.onUpdate}
                        onComplete={this.onComplete}
                        extensions={this.props.extensions}
                        disabled={isExtensionEnabled(this.props.configurationCascade.merged, this.props.extension.id)}
                    >
                        Enable extension
                    </ExtensionConfigureDropdownItem>
                    <ExtensionConfigureDropdownItem
                        extension={this.props.extension}
                        subject={this.props.subject}
                        operation={this.disableExtensionForSubject}
                        onUpdate={this.props.onUpdate}
                        onComplete={this.onComplete}
                        extensions={this.props.extensions}
                        disabled={!isExtensionEnabled(this.props.configurationCascade.merged, this.props.extension.id)}
                    >
                        Disable extension
                    </ExtensionConfigureDropdownItem>
                    {// Hide "Remove extension" button when the extension is present in other lower-precedence
                    // subjects' settings, because in that case, removing the extension from user settings
                    // would just fall back to the lower-precedence settings, which would be unexpected to the
                    // user. To handle these cases, the user must manually edit settings.
                    otherSubjectsWithExtensionAdded.length === 0 ? (
                        <ExtensionConfigureDropdownItem
                            extension={this.props.extension}
                            subject={this.props.subject}
                            operation={this.removeExtensionForSubject}
                            onUpdate={this.props.onUpdate}
                            onComplete={this.onComplete}
                            extensions={this.props.extensions}
                        >
                            Remove extension
                        </ExtensionConfigureDropdownItem>
                    ) : (
                        <>
                            <DropdownItem divider={true} />
                            <DropdownItem
                                header={true}
                                title={otherSubjectsWithExtensionAdded
                                    .filter(({ subject }) => subject.__typename === 'Org')
                                    .map(({ subject }) => subjectLabel(subject))
                                    .join(', ')}
                            >
                                <this.props.extensions.context.icons.Info className="icon-inline mr-1" />
                                {otherSubjectsWithExtensionAdded.some(({ subject }) => subject.__typename === 'Site')
                                    ? 'Default: enabled for everyone'
                                    : 'Default: enabled for organization'}
                            </DropdownItem>
                        </>
                    )}
                </DropdownMenu>
            </ButtonDropdown>
        )
    }

    private toggle = () => {
        this.setState(prevState => ({ dropdownOpen: !prevState.dropdownOpen }))
    }

    private enableExtensionForSubject = (
        extension: ConfiguredExtension,
        subject: ConfiguredSubjectOrError<ConfigurationSubject, Settings>
    ) =>
        this.props.extensions.context.updateExtensionSettings(subject.subject.id, {
            extensionID: extension.id,
            enabled: true,
        })

    private disableExtensionForSubject = (
        extension: ConfiguredExtension,
        subject: ConfiguredSubjectOrError<ConfigurationSubject, Settings>
    ) =>
        this.props.extensions.context.updateExtensionSettings(subject.subject.id, {
            extensionID: extension.id,
            enabled: false,
        })

    private removeExtensionForSubject = (
        extension: ConfiguredExtension,
        subject: ConfiguredSubjectOrError<ConfigurationSubject, Settings>
    ) =>
        this.props.extensions.context.updateExtensionSettings(subject.subject.id, {
            extensionID: extension.id,
            remove: true,
        })

    private onComplete = () => this.setState({ dropdownOpen: false })
}
