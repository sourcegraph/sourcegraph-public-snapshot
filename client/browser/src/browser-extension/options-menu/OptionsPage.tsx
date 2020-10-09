import React, { useCallback, useMemo, useState } from 'react'
import { useInputValidation, deriveInputClassName } from '../../../../shared/src/util/useInputValidation'
import { LoaderInput } from '../../../../shared/src/components/LoaderInput'
import BookOpenPageVariantIcon from 'mdi-react/BookOpenPageVariantIcon'
import classNames from 'classnames'
import { LinkOrButton } from '../../../../shared/src/components/LinkOrButton'
import { Observable } from 'rxjs'
import { Toggle } from '../../../../shared/src/components/Toggle'
import { SourcegraphLogo } from './SourcegraphLogo'
import { noop } from 'lodash'
import { OptionsPageAdvancedSettings } from './OptionsPageAdvancedSettings'
import EarthIcon from 'mdi-react/EarthIcon'
import LockIcon from 'mdi-react/LockIcon'
import { knownCodeHosts } from '../knownCodeHosts'

interface OptionsPageProps {
    version: string
    sourcegraphUrl: string
    isActivated: boolean
    onToggleActivated: (value: boolean) => void
    validateSourcegraphUrl: (url: string) => Observable<string | undefined>
    isFullPage: boolean
    showPrivateRepositoryAlert?: boolean
    permissionAlert?: { name: string; icon?: JSX.Element }
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
                asynchronousValidators: [validateSourcegraphUrl],
            }),
            [sourcegraphUrl, validateSourcegraphUrl]
        )
    )

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
                <form>
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
                <p>
                    <LinkOrButton>How do we keep your code private?</LinkOrButton>
                </p>
            </section>

            {permissionAlert && <PermissionAlert {...permissionAlert} onClickGrantPermissions={noop} />}

            {showPrivateRepositoryAlert && <PrivateRepositoryAlert />}
            <section className="options-page__section">
                <p>
                    <LinkOrButton
                        onSelect={useCallback(() => setShowAdvancedSettings(!showAdvancedSettings), [
                            showAdvancedSettings,
                        ])}
                    >
                        {showAdvancedSettings ? 'Hide' : 'Show'} advanced settings
                    </LinkOrButton>
                </p>
            </section>
            {showAdvancedSettings && (
                <OptionsPageAdvancedSettings optionFlags={optionFlags} onChangeOptionFlag={onChangeOptionFlag} />
            )}
            <section className="options-page__split-section">
                <div className="options-page__split-section__part">
                    <LinkOrButton to="https://sourcegraph.com">
                        <EarthIcon className="icon-inline" /> Sourcegraph Cloud
                    </LinkOrButton>
                </div>
                <div className="options-page__split-section__part">
                    <LinkOrButton to="https://docs.sourcegraph.com">
                        <BookOpenPageVariantIcon className="icon-inline" /> Documentation
                    </LinkOrButton>
                </div>
            </section>
        </div>
    )
}

interface PermissionAlertProps {
    icon?: JSX.Element
    name: string
    onClickGrantPermissions: () => void
}

const PermissionAlert: React.FunctionComponent<PermissionAlertProps> = ({ name, icon, onClickGrantPermissions }) => (
    <section className="options-page__section options-page__alert">
        <h4>
            {icon} {name}
        </h4>
        <p>
            <strong>Grant the permissions</strong> to use the Sourcegraph extension on {name}.
        </p>
        <button onClick={onClickGrantPermissions} className="btn btn-primary">
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
    <section className="options-page__section options-page__alert">
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
