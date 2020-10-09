import React, { useCallback, useMemo, useState } from 'react'
import { useInputValidation, deriveInputClassName } from '../../../../shared/src/util/useInputValidation'
import { LoaderInput } from '../../../../branded/src/components/LoaderInput'
import BookOpenPageVariantIcon from 'mdi-react/BookOpenPageVariantIcon'
import classNames from 'classnames'
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'
import { Observable } from 'rxjs'
import { Toggle } from '../../../../branded/src/components/Toggle'
import { SourcegraphLogo } from './SourcegraphLogo'
import { OptionsPageAdvancedSettings } from './OptionsPageAdvancedSettings'
import EarthIcon from 'mdi-react/EarthIcon'
import LockIcon from 'mdi-react/LockIcon'
import CheckCircleOutlineIcon from 'mdi-react/CheckCircleOutlineIcon'
import { knownCodeHosts } from '../knownCodeHosts'

export interface OptionsPageProps {
    version: string
    sourcegraphUrl: string
    isActivated: boolean
    onToggleActivated: (value: boolean) => void
    validateSourcegraphUrl: (url: string) => Observable<string | undefined>
    isFullPage: boolean
    showPrivateRepositoryAlert?: boolean
    showSourcegraphCloudAlert?: boolean
    permissionAlert?: { name: string; icon?: React.ComponentType<{ className?: string }> }
    requestPermissionsHandler?: React.MouseEventHandler
    optionFlags: { key: string; label: string; value: boolean }[]
    onChangeOptionFlag: (key: string, value: boolean) => void
    currentHost?: string
}

export const OptionsPage: React.FunctionComponent<OptionsPageProps> = ({
    version,
    sourcegraphUrl,
    validateSourcegraphUrl,
    isActivated,
    onToggleActivated,
    isFullPage,
    showPrivateRepositoryAlert,
    showSourcegraphCloudAlert,
    permissionAlert,
    requestPermissionsHandler,
    optionFlags,
    onChangeOptionFlag,
    currentHost,
}) => {
    const [showAdvancedSettings, setShowAdvancedSettings] = useState(false)
    const [urlState, nextUrlFieldChange, urlInputReference] = useInputValidation(
        useMemo(
            () => ({
                initialValue: sourcegraphUrl,
                synchronousValidators: [onlyHTTPS],
                asynchronousValidators: [validateSourcegraphUrl],
            }),
            [sourcegraphUrl, validateSourcegraphUrl]
        )
    )

    const toggleAdvancedSettings = useCallback(
        () => setShowAdvancedSettings(showAdvancedSettings => !showAdvancedSettings),
        []
    )

    const linkProps: Pick<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'rel' | 'target'> = {
        target: '_blank',
        rel: 'noopener noreferrer',
    }

    return (
        <div className={classNames('options-page', { 'options-page--full': isFullPage })}>
            <section className="options-page__section">
                <div className="d-flex justify-content-between">
                    <SourcegraphLogo className="options-page__logo" />
                    <div>
                        <Toggle
                            value={isActivated}
                            onToggle={onToggleActivated}
                            title={`Toggle to ${isActivated ? 'disable' : 'enable'} extension`}
                        />
                    </div>
                </div>
                <div className="options-page__version">v{version}</div>
            </section>
            <CodeHostsSection currentHost={currentHost} />
            <section className="options-page__section">
                {/* eslint-disable-next-line react/forbid-elements */}
                <form onSubmit={preventDefault} noValidate={true}>
                    <label htmlFor="sourcegraph-url">Sourcegraph URL</label>
                    <LoaderInput loading={urlState.loading} className={classNames(deriveInputClassName(urlState))}>
                        <input
                            className="form-control"
                            id="sourcegraph-url"
                            type="url"
                            pattern="^https://.*"
                            value={urlState.value}
                            onChange={nextUrlFieldChange}
                            ref={urlInputReference}
                        />
                    </LoaderInput>
                    {urlState.loading ? (
                        <small className="text-muted d-block mt-1">Checking...</small>
                    ) : urlState.kind === 'INVALID' ? (
                        <small className="invalid-feedback">{urlState.reason}</small>
                    ) : (
                        <small className="valid-feedback">Looks good!</small>
                    )}
                </form>
                <p className="mt-3 mb-1">
                    <small>Enter the URL of your Sourcegraph instance to use the extension on private code.</small>
                </p>

                <a href="https://docs.sourcegraph.com/integration/browser_extension#privacy" {...linkProps}>
                    <small>How do we keep your code private?</small>
                </a>
            </section>

            {permissionAlert && (
                <PermissionAlert {...permissionAlert} onClickGrantPermissions={requestPermissionsHandler} />
            )}

            {showSourcegraphCloudAlert && <SourcegraphCloudAlert />}

            {showPrivateRepositoryAlert && <PrivateRepositoryAlert />}
            <section className="options-page__section pt-2">
                <p className="mb-0">
                    <button type="button" className="btn btn-link btn-sm p-0" onClick={toggleAdvancedSettings}>
                        <small>{showAdvancedSettings ? 'Hide' : 'Show'} advanced settings</small>
                    </button>
                </p>
                {showAdvancedSettings && (
                    <OptionsPageAdvancedSettings optionFlags={optionFlags} onChangeOptionFlag={onChangeOptionFlag} />
                )}
            </section>
            <section className="d-flex">
                <div className="options-page__split-section-part">
                    <a href="https://sourcegraph.com/search" {...linkProps}>
                        <EarthIcon className="icon-inline mr-2" />
                        Sourcegraph Cloud
                    </a>
                </div>
                <div className="options-page__split-section-part" {...linkProps}>
                    <a href="https://docs.sourcegraph.com">
                        <BookOpenPageVariantIcon className="icon-inline mr-2" />
                        Documentation
                    </a>
                </div>
            </section>
        </div>
    )
}

interface PermissionAlertProps {
    icon?: React.ComponentType<{ className?: string }>
    name: string
    onClickGrantPermissions?: React.MouseEventHandler
}

const PermissionAlert: React.FunctionComponent<PermissionAlertProps> = ({
    name,
    icon: Icon,
    onClickGrantPermissions,
}) => (
    <section className="options-page__section options-page__alert">
        <h6 className="d-flex align-items-center">
            {Icon && <Icon className="icon-inline mr-2" />} <span>{name}</span>
        </h6>
        <p className="options-page__permission-text">
            <strong>Grant the permissions</strong> to use the Sourcegraph extension on {name}.
        </p>
        <button type="button" onClick={onClickGrantPermissions} className="btn btn-sm btn-primary">
            <small>Grant permissions</small>
        </button>
    </section>
)

const PrivateRepositoryAlert: React.FunctionComponent = () => (
    <section className="options-page__section options-page__alert">
        <h6 className="d-flex align-items-center">
            <LockIcon className="icon-inline mr-2" />
            Private repository
        </h6>
        <p>
            To use the browser extension with your private repositories, you need to set up a{' '}
            <strong>private Sourcegraph instance</strong> and connect it to the extension.
        </p>
        <ol>
            <li>
                <a href="#">Install and configure Sourcegraph</a>. Skip this step if you already have a private
                Sourcegraph instance.
            </li>
            <li>Click the Sourcegraph extension icon in the browser toolbar to open the settings page</li>
            <li>
                Enter the URL (including the protocol) of your Sourcegraph instance (such as
                https://sourcegraph.example.com above).
            </li>
            <li>Make sure that the status shows 'connected'.</li>
        </ol>
    </section>
)

const CodeHostsSection: React.FunctionComponent<{ currentHost?: string }> = ({ currentHost }) => (
    <section className="options-page__section options-page__code-hosts">
        <p>Get code intelligence tooltips while browsing files and reading PRs on your code host.</p>
        <div>
            {knownCodeHosts.map(({ host, icon: Icon }) => (
                <span
                    key={host}
                    className={classNames('code-hosts-section__icon', {
                        // Use `endsWith` in order to match subdomains.
                        'code-hosts-section__icon--highlighted': currentHost?.endsWith(host),
                    })}
                >
                    {Icon && <Icon />}
                </span>
            ))}
        </div>
    </section>
)

const SourcegraphCloudAlert: React.FunctionComponent = () => (
    <section className="options-page__section options-page__alert">
        <h6 className="d-flex align-items-center">
            <CheckCircleOutlineIcon className="icon-inline mr-2" />
            You're on Sourcegraph Cloud
        </h6>
        <p>Naturally, browser extension is not necessary to browse public code on Sourcegraph.com</p>
    </section>
)

function preventDefault(event: React.FormEvent<HTMLFormElement>): void {
    event.preventDefault()
}

/**
 * Synchronous validator to provide helpful error message
 */
function onlyHTTPS(url: string): string | undefined {
    // TODO(tj): improve copy
    return url.startsWith('https://') ? undefined : 'We support only https'
}
