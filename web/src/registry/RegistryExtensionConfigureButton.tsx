import WarningIcon from '@sourcegraph/icons/lib/Warning'
import { upperFirst } from 'lodash'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { catchError, map, mapTo, startWith, switchMap, tap } from 'rxjs/operators'
import { currentUser } from '../auth'
import * as GQL from '../backend/graphqlschema'
import { asError, ErrorLike, isErrorLike } from '../util/errors'
import { updateUserExtensionSettings } from './backend'

/** The extension can be specified by providing the GQL.IRegistryExtension or just by the extension ID. */
type ExtensionSpec =
    | {
          extensionGQLID: GQL.ID
          extensionID?: string
      }
    | { extensionGQLID?: null; extensionID: string }

type Props = ExtensionSpec & {
    /** The subject whose enablement state is reflected in this component. */
    subject: GQL.ID

    /**
     * Also show a button to remove (not just enable/disable) the extension in settings. This is used when the
     * settings refer to an invalid extension and the user wants to remove it from settings entirely.
     */
    showRemove?: boolean

    viewerCanConfigure: boolean
    isEnabled: boolean

    compact?: boolean

    className?: string
    buttonClassName?: string
    disabled?: boolean

    /** Called when the extension is enabled or disabled. */
    onDidUpdate: () => void
}

interface State {
    /** Undefined means in progress, null means done or not started. */
    configureOrError?: null | ErrorLike

    currentUserSubject?: GQL.ID | null
}

/** A button that enables/disables/removes an extension in user settings. */
export class RegistryExtensionConfigureButton extends React.PureComponent<Props, State> {
    public state: State = {
        configureOrError: null,
    }

    private settingsUpdates = new Subject<
        Pick<GQL.IUpdateExtensionOnConfigurationMutationArguments, 'enabled' | 'remove'>
    >()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(currentUser.subscribe(user => this.setState({ currentUserSubject: user && user.id })))

        this.subscriptions.add(
            this.settingsUpdates
                .pipe(
                    switchMap(args =>
                        updateUserExtensionSettings({
                            extension: this.props.extensionGQLID,
                            extensionID: this.props.extensionID,
                            ...args,
                        }).pipe(
                            mapTo(null),
                            catchError(error => [asError(error)]),
                            map(c => ({ configureOrError: c })),
                            tap(() => {
                                if (this.props.onDidUpdate) {
                                    this.props.onDidUpdate()
                                }
                            }),
                            startWith<Pick<State, 'configureOrError'>>({
                                configureOrError: undefined,
                            })
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
        const buttonClassName = `${this.props.buttonClassName || 'btn-sm'} ${this.props.compact ? 'py-0' : ''}`

        const subjectIsSelf = this.state.currentUserSubject && this.state.currentUserSubject === this.props.subject

        if (!this.props.viewerCanConfigure || !subjectIsSelf) {
            return (
                <small
                    className={`registry-extension-configure-button d-inline-block ${this.props.className || ''} ${
                        this.props.isEnabled ? 'text-success' : 'text-muted'
                    }`}
                    data-tooltip={!subjectIsSelf ? 'Edit settings to enable/disable this extension' : undefined}
                >
                    {this.props.isEnabled ? 'Enabled' : 'Disabled'}
                </small>
            )
        }

        return (
            <div className={`registry-extension-configure-button btn-group ${this.props.className || ''}`} role="group">
                {this.props.showRemove &&
                    subjectIsSelf && (
                        <button
                            className={`registry-extension-configure-button__btn btn btn-secondary mr-2 ${buttonClassName}`}
                            onClick={this.removeExtensionSettings}
                            disabled={this.props.disabled || this.state.configureOrError === undefined}
                            title={subjectIsSelf ? `Remove extension from settings` : undefined}
                        >
                            Remove {!this.props.compact && ' from settings'}
                        </button>
                    )}
                <button
                    className={`registry-extension-configure-button__btn btn registry-extension-configure-button__enable-btn btn-${
                        this.props.isEnabled ? 'link' : 'success'
                    } ${buttonClassName}`}
                    onClick={this.toggleExtensionEnabled}
                    disabled={this.props.disabled || this.state.configureOrError === undefined}
                    title={subjectIsSelf ? `${this.props.isEnabled ? 'Disable' : 'Enable'} extension` : undefined}
                >
                    {this.props.isEnabled ? 'Disable' : 'Enable'}
                    {!this.props.compact && ' extension'}
                </button>
                {isErrorLike(this.state.configureOrError) && (
                    <button
                        disabled={true}
                        className={`btn btn-danger ${buttonClassName}`}
                        title={upperFirst(this.state.configureOrError.message)}
                    >
                        <WarningIcon className="icon-inline" />
                    </button>
                )}
            </div>
        )
    }

    private toggleExtensionEnabled = () => this.settingsUpdates.next({ enabled: !this.props.isEnabled })
    private removeExtensionSettings = () => this.settingsUpdates.next({ remove: true })
}
