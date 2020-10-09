import React, { useCallback, useMemo, useState } from 'react'
import { useInputValidation, deriveInputClassName } from '../../../../shared/src/util/useInputValidation'
import { LoaderInput } from '../../../../branded/src/components/LoaderInput'
import BookOpenPageVariantIcon from 'mdi-react/BookOpenPageVariantIcon'
import classNames from 'classnames'
import { ButtonLink } from '../../../../shared/src/components/LinkOrButton'
import { Observable } from 'rxjs'
import { Toggle } from '../../../../branded/src/components/Toggle'
import { SourcegraphLogo } from './SourcegraphLogo'
import { noop } from 'lodash'
import { OptionsPageAdvancedSettings } from './OptionsPageAdvancedSettings'
import EarthIcon from 'mdi-react/EarthIcon'
import LockIcon from 'mdi-react/LockIcon'
import { knownCodeHosts } from '../knownCodeHosts'
import { MdiReactIconProps } from 'mdi-react'

export interface OptionsPageProps {
    version: string
    sourcegraphUrl: string
    isActivated: boolean
    onToggleActivated: (value: boolean) => void
    validateSourcegraphUrl: (url: string) => Observable<string | undefined>
    isFullPage: boolean
    showPrivateRepositoryAlert?: boolean
    permissionAlert?: { name: string; icon?: React.ComponentType<MdiReactIconProps> }
    optionFlags: { key: string; label: string; value: boolean }[]
    onChangeOptionFlag: (key: string, value: boolean) => void
    currentHost?: string
}

function onlyHTTPS(url: string): string | undefined {
    // TODO(tj): improve copy
    return url.startsWith('https://') ? undefined : 'We support only https'
}

export const OptionsPage: React.FunctionComponent<OptionsPageProps> = ({
    version,
    sourcegraphUrl,
    validateSourcegraphUrl,
    isActivated,
    onToggleActivated,
    isFullPage,
    showPrivateRepositoryAlert,
    permissionAlert,
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

    const toggleAdvancedSettings = useCallback(() => setShowAdvancedSettings(showAdvancedSettings => !showAdvancedSettings), [])

    const linkProps: Pick<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'rel' | 'target'> = {
        target: '_blank',
        rel: 'noopener noreferrer'
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
                            className="form-control options-page__input"
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
                <p className="mt-2">Enter the URL of your Sourcegraph instance to use the extension on private code.</p>

                    <ButtonLink to="https://docs.sourcegraph.com/integration/browser_extension#privacy" {...linkProps}>
                        How do we keep your code private?
                        </ButtonLink>

            </section>

            {permissionAlert && <PermissionAlert {...permissionAlert} onClickGrantPermissions={noop} />}

            {showPrivateRepositoryAlert && <PrivateRepositoryAlert />}
            <section className="options-page__section">
                <p className="mb-0">
                    <ButtonLink
                        onSelect={toggleAdvancedSettings}
                    >
                        {showAdvancedSettings ? 'Hide' : 'Show'} advanced settings
                    </ButtonLink>
                </p>
            {showAdvancedSettings && (
                <OptionsPageAdvancedSettings optionFlags={optionFlags} onChangeOptionFlag={onChangeOptionFlag} />
                )}
                </section>
            <section className="options-page__split-section">
                <div className="options-page__split-section__part">
                    <ButtonLink to="https://sourcegraph.com/search" {...linkProps}>
                        <EarthIcon className="icon-inline mr-2" />Sourcegraph Cloud
                    </ButtonLink>
                </div>
                <div className="options-page__split-section__part">
                    <ButtonLink to="https://docs.sourcegraph.com">
                        <BookOpenPageVariantIcon className="icon-inline mr-2" {...linkProps} />Documentation
                    </ButtonLink>
                </div>
            </section>
        </div>
    )
}

interface PermissionAlertProps {
    icon?: React.ComponentType<MdiReactIconProps>
    name: string
    onClickGrantPermissions: () => void
}

const PermissionAlert: React.FunctionComponent<PermissionAlertProps> = ({ name, icon: Icon, onClickGrantPermissions }) => (
    <section className="options-page__section options-page__alert">
        <h4>
            {Icon && <Icon className="icon-inline" />} {name}
        </h4>
        <p className="options-page__permission-text">
            <strong>Grant the permissions</strong> to use the Sourcegraph extension on {name}.
        </p>
        <button type="button" onClick={onClickGrantPermissions} className="btn btn-primary">
            Grant permissions
        </button>
    </section>
)

const PrivateRepositoryAlert: React.FunctionComponent = props => (
    <section className="options-page__section options-page__alert">
        <h3>
            <LockIcon className="icon-inline" />
            Private repository
        </h3>
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

function preventDefault(event: React.FormEvent<HTMLFormElement>): void {
    event.preventDefault()
}
