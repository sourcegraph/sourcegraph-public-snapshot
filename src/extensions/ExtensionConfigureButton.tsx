import * as React from 'react'
import { Link } from 'react-router-dom'
import { ButtonDropdown, DropdownMenu, DropdownToggle } from 'reactstrap'
import DropdownItem from 'reactstrap/lib/DropdownItem'
import { from, Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { ExtensionsProps } from '../context'
import { asError, ErrorLike, isErrorLike } from '../errors'
import {
    ConfigurationCascadeProps,
    ConfigurationSubject,
    ConfiguredSubjectOrError,
    Settings,
    SUBJECT_TYPE_ORDER,
    subjectLabel,
    subjectTypeHeader,
} from '../settings'
import { ConfiguredExtension, isExtensionAdded, isExtensionEnabled } from './extension'

interface ExtensionConfiguredSubject {
    extension: ConfiguredExtension
    subject: ConfiguredSubjectOrError<ConfigurationSubject, Settings>
}

/** A dropdown menu item for a extension-subject item that links to the subject's settings.  */
export class ExtensionConfiguredSubjectItemForConfigure<
    S extends ConfigurationSubject,
    C extends Settings
> extends React.PureComponent<
    {
        item: ExtensionConfiguredSubject
        onUpdate: () => void
        onComplete: () => void
    } & ExtensionsProps<S, C>
> {
    public render(): JSX.Element | null {
        return (
            <Link
                className="dropdown-item d-flex justify-content-between align-items-center"
                to={this.props.item.subject.subject.settingsURL}
            >
                <span className="mr-4">{subjectLabel(this.props.item.subject.subject)}</span>
                {isExtensionAdded(this.props.item.subject.settings, this.props.item.extension.id) &&
                    !isErrorLike(this.props.item.subject.settings) &&
                    !isExtensionEnabled(this.props.item.subject.settings, this.props.item.extension.id) && (
                        <small className="text-muted">Disabled</small>
                    )}
            </Link>
        )
    }
}

const LOADING: 'loading' = 'loading'

interface ExtensionConfiguredSubjectItemForAddState {
    /** The add operation's status: null when done or not started, 'loading', or an error. */
    addOrError: typeof LOADING | null | ErrorLike
}

/** A dropdown menu item for a extension-subject item that adds the extension to the subject's settings.  */
export class ExtensionConfiguredSubjectItemForAdd<
    S extends ConfigurationSubject,
    C extends Settings
> extends React.PureComponent<
    {
        item: ExtensionConfiguredSubject
        confirm?: () => boolean
        onUpdate: () => void
        onComplete: () => void
    } & ExtensionsProps<S, C>,
    ExtensionConfiguredSubjectItemForAddState
> {
    public state: ExtensionConfiguredSubjectItemForAddState = { addOrError: null }

    private addClicks = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.addClicks
                .pipe(
                    switchMap(() =>
                        from(
                            this.props.extensions.context.updateExtensionSettings(this.props.item.subject.subject.id, {
                                extensionID: this.props.item.extension.id,
                                enabled: true,
                            })
                        ).pipe(
                            mapTo(null),
                            tap(() => this.props.onComplete()),
                            catchError(error => [asError(error) as ErrorLike]),
                            map(c => ({ addOrError: c } as ExtensionConfiguredSubjectItemForAddState)),
                            tap(() => this.props.onUpdate()),
                            startWith<ExtensionConfiguredSubjectItemForAddState>({ addOrError: LOADING })
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
        const isAdded = isExtensionAdded(this.props.item.subject.settings, this.props.item.extension.id)
        return (
            <DropdownItem
                className="dropdown-item d-flex justify-content-between align-items-center"
                disabled={isAdded}
                onClick={this.onClick}
                toggle={false}
            >
                <span className="mr-5">{subjectLabel(this.props.item.subject.subject)}</span>
                <div>
                    {isErrorLike(this.state.addOrError) && (
                        <small className="text-danger" title={this.state.addOrError.message}>
                            <this.props.extensions.context.icons.Warning className="icon-inline" /> Error
                        </small>
                    )}
                    {isAdded && <span className="small text-muted">Already added</span>}
                </div>
            </DropdownItem>
        )
    }

    private onClick: React.MouseEventHandler<HTMLElement> = () => {
        if (!this.props.confirm || this.props.confirm()) {
            this.addClicks.next()
        }
    }
}

interface ExtensionConfiguredSubjectItemForRemoveState {
    /** The remove operation's status: null when done or not started, 'loading', or an error. */
    removeOrError: typeof LOADING | null | ErrorLike
}

/** A dropdown menu item for a extension-subject item that removes the extension from the subject's settings.  */
export class ExtensionConfiguredSubjectItemForRemove<
    S extends ConfigurationSubject,
    C extends Settings
> extends React.PureComponent<
    {
        item: ExtensionConfiguredSubject
        confirm?: () => boolean
        onUpdate: () => void
        onComplete: () => void
    } & ExtensionsProps<S, C>,
    ExtensionConfiguredSubjectItemForRemoveState
> {
    public state: ExtensionConfiguredSubjectItemForRemoveState = { removeOrError: null }

    private removeClicks = new Subject<void>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.removeClicks
                .pipe(
                    switchMap(() =>
                        from(
                            this.props.extensions.context.updateExtensionSettings(this.props.item.subject.subject.id, {
                                extensionID: this.props.item.extension.id,
                                remove: true,
                            })
                        ).pipe(
                            mapTo(null),
                            tap(() => this.props.onComplete()),
                            catchError(error => [asError(error) as ErrorLike]),
                            map(c => ({ removeOrError: c } as ExtensionConfiguredSubjectItemForRemoveState)),
                            tap(() => this.props.onUpdate()),
                            startWith<ExtensionConfiguredSubjectItemForRemoveState>({ removeOrError: LOADING })
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
                onClick={this.onClick}
                toggle={false}
            >
                <span className="mr-5">{subjectLabel(this.props.item.subject.subject)}</span>
                <div>
                    {isErrorLike(this.state.removeOrError) && (
                        <small className="text-danger" title={this.state.removeOrError.message}>
                            <this.props.extensions.context.icons.Warning className="icon-inline" /> Error
                        </small>
                    )}
                </div>
            </DropdownItem>
        )
    }

    private onClick: React.MouseEventHandler<HTMLElement> = () => {
        if (!this.props.confirm || this.props.confirm()) {
            this.removeClicks.next()
        }
    }
}

class ExtensionConfigurationSubjectsDropdownItems<
    S extends ConfigurationSubject,
    C extends Settings
> extends React.PureComponent<
    {
        items: ExtensionConfiguredSubject[]

        /**
         * Closes the dropdown menu. This is necessary because the menu must remain open after the user selects an
         * item that starts an operation. If it immediately closed, then the component's componentWillUnmount would
         * be called and the in-progress operation would be canceled (i.e., the HTTP request would be canceled,
         * probably before it reached the server). Also, if the operation failed, the user would not get any
         * feedback about the error (because it is shown in the menu item).
         */
        onComplete: () => void
    } & Pick<Props<S, C>, 'header' | 'itemComponent' | 'confirm' | 'onUpdate'> &
        ExtensionsProps<S, C>
> {
    public render(): JSX.Element | null {
        const { header, items, itemComponent: Item, ...props } = this.props

        const itemsByType = new Map<
            ExtensionConfiguredSubject['subject']['subject']['__typename'],
            ExtensionConfiguredSubject[]
        >()
        for (const item of items) {
            let typeItems = itemsByType.get(item.subject.subject.__typename)
            if (!typeItems) {
                typeItems = []
                itemsByType.set(item.subject.subject.__typename, typeItems)
            }
            typeItems.push(item)
        }
        let needsDivider = false
        return (
            <>
                {header && (
                    <>
                        <DropdownItem header={true}>{header}</DropdownItem>
                        <DropdownItem divider={true} />
                    </>
                )}
                {SUBJECT_TYPE_ORDER.map((nodeType, i) => {
                    const items = itemsByType.get(nodeType)
                    if (!items) {
                        return null
                    }
                    const neededDivider = needsDivider
                    needsDivider = items.length > 0
                    const headerLabel = subjectTypeHeader(nodeType)
                    return (
                        <React.Fragment key={i}>
                            {neededDivider && <DropdownItem divider={true} />}
                            {headerLabel && <DropdownItem header={true}>{headerLabel}</DropdownItem>}
                            {items.map((item, i) => (
                                <Item key={i} item={item} {...props} />
                            ))}
                        </React.Fragment>
                    )
                })}
            </>
        )
    }
}

interface ExtensionConfigurationSubjectsFilter {
    added?: boolean
    notAdded?: boolean
    onlyIfViewerCanAdminister?: boolean
}

export const ALL_CAN_ADMINISTER: ExtensionConfigurationSubjectsFilter = {
    added: true,
    notAdded: true,
    onlyIfViewerCanAdminister: true,
}

export const ADDED_AND_CAN_ADMINISTER: ExtensionConfigurationSubjectsFilter = {
    added: true,
    notAdded: false,
    onlyIfViewerCanAdminister: true,
}

export function filterItems<S extends ConfigurationSubject>(
    extensionID: string,
    items: ConfiguredSubjectOrError<S>[],
    filter: ExtensionConfigurationSubjectsFilter
): ConfiguredSubjectOrError<S>[] {
    return items.filter(item => {
        const isAdded = isExtensionAdded(item.settings, extensionID)
        if (isAdded && !filter.added) {
            return false
        }
        if (!isAdded && !filter.notAdded) {
            return false
        }
        if (!item.subject.viewerCanAdminister && filter.onlyIfViewerCanAdminister) {
            return false
        }
        return true
    })
}

interface Props<S extends ConfigurationSubject, C extends Settings>
    extends ConfigurationCascadeProps<S, C>,
        ExtensionsProps<S, C> {
    /** The extension that this element is for. */
    extension: ConfiguredExtension

    /** Class name applied to the button element. */
    buttonClassName?: string

    /* The button label. */
    children: React.ReactFragment

    /** The dropdown menu header. */
    header?: React.ReactFragment

    /** Only show items matching the filter. */
    itemFilter: ExtensionConfigurationSubjectsFilter

    /** Renders the subject dropdown item. */
    itemComponent: React.ComponentType<
        {
            item: ExtensionConfiguredSubject
            onUpdate: () => void
            onComplete: () => void
        } & ExtensionsProps<S, C>
    >

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
 * Displays a button with a dropdown menu listing the extension configuration subjects of an extension.
 */
export class ExtensionConfigureButton<S extends ConfigurationSubject, C extends Settings> extends React.PureComponent<
    Props<S, C>,
    State
> {
    public state: State = {
        dropdownOpen: false,
    }

    public render(): JSX.Element | null {
        if (!this.props.configurationCascade.subjects) {
            return null
        }
        if (isErrorLike(this.props.configurationCascade.subjects)) {
            // TODO: Show error.
            return null
        }
        const configurableSubjects = filterItems(
            this.props.extension.id,
            this.props.configurationCascade.subjects,
            this.props.itemFilter
        )
        return (
            <ButtonDropdown isOpen={this.state.dropdownOpen} toggle={this.toggle} group={this.props.caret !== false}>
                <DropdownToggle
                    caret={this.props.caret !== false}
                    color=""
                    className={this.props.buttonClassName}
                    disabled={configurableSubjects.length === 0}
                >
                    {this.props.children}
                </DropdownToggle>
                <DropdownMenu>
                    <ExtensionConfigurationSubjectsDropdownItems
                        header={this.props.header}
                        itemComponent={this.props.itemComponent}
                        items={configurableSubjects.map(subject => ({
                            subject,
                            extension: this.props.extension,
                        }))}
                        confirm={this.props.confirm}
                        onUpdate={this.props.onUpdate}
                        onComplete={this.onComplete}
                        extensions={this.props.extensions}
                    />
                </DropdownMenu>
            </ButtonDropdown>
        )
    }

    private toggle = () => {
        this.setState(prevState => ({ dropdownOpen: !prevState.dropdownOpen }))
    }

    private onComplete = () => this.setState({ dropdownOpen: false })
}
