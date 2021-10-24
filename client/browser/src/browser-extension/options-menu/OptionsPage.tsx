import classNames from 'classnames'
import CheckCircleOutlineIcon from 'mdi-react/CheckCircleOutlineIcon'
import LockIcon from 'mdi-react/LockIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import React, { useCallback, useState } from 'react'
import { of } from 'rxjs'

import { SourcegraphLogo } from '@sourcegraph/branded/src/components/SourcegraphLogo'
import { Toggle } from '@sourcegraph/branded/src/components/Toggle'

import { CLOUD_SOURCEGRAPH_URL, isCloudSourcegraphUrl } from '../../shared/util/context'

import { OptionsPageAdvancedSettings } from './components/OptionsPageAdvancedSettings'
import { SourcegraphURLInput, SourcegraphURLInputProps } from './components/SourcegraphUrlInput'
import { LINK_PROPS } from './constants'

interface PermissionAlertProps {
    icon?: React.ComponentType<{ className?: string }>
    name: string
    onClickGrantPermissions?: React.MouseEventHandler
}

const PermissionAlert: React.FC<PermissionAlertProps> = ({ name, icon: Icon, onClickGrantPermissions }) => (
    <section className="options-page__section bg-2">
        <h4>
            {Icon && <Icon className="icon-inline mr-2" />} <span>{name}</span>
        </h4>
        <p className="options-page__permission-text">
            <strong>Grant permissions</strong> to use the Sourcegraph extension on {name}.
        </p>
        <button type="button" onClick={onClickGrantPermissions} className="btn btn-sm btn-primary">
            <small>Grant permissions</small>
        </button>
    </section>
)

// TODO: mention CLOUD instead
const PrivateRepositoryAlert: React.FC = () => (
    <section className="options-page__section bg-2">
        <h4>
            <LockIcon className="icon-inline mr-2" />
            Private repository
        </h4>
        <p>
            To use the browser extension with your private repositories, you need to set up a{' '}
            <strong>private Sourcegraph instance</strong> and connect the browser extension to it.
        </p>
        <ol>
            <li className="mb-2">
                <a href="https://docs.sourcegraph.com/" rel="noopener" target="_blank">
                    Install and configure Sourcegraph
                </a>
                . Skip this step if you already have a private Sourcegraph instance.
            </li>
            <li className="mb-2">
                Enter the URL (including the protocol) of your Sourcegraph instance above, e.g.{' '}
                <q>https://sourcegraph.example.com</q>.
            </li>
            <li>
                Make sure that the status says <q>Looks good!</q>.
            </li>
        </ol>
    </section>
)

const InfoSection: React.FC = () => (
    <section className="options-page__section">
        Get code intelligence tooltips while browsing and reviewing code on your code host.{' '}
        <a href="https://docs.sourcegraph.com/integration/browser_extension" {...LINK_PROPS}>
            Learn more
        </a>{' '}
        Learn more about the extension and compatible code hosts.
    </section>
)

const SourcegraphCloudAlert: React.FC = () => (
    <section className="options-page__section bg-2">
        <h4>
            <CheckCircleOutlineIcon className="icon-inline mr-2" />
            You're on Sourcegraph Cloud
        </h4>
        <p>Naturally, the browser extension is not necessary to browse public code on sourcegraph.com.</p>
    </section>
)

export interface OptionsPageProps {
    version: string

    // Self hosted Sourcegraph URL
    selfHostedSourcegraphURL?: string
    validateSourcegraphUrl: SourcegraphURLInputProps['validate']
    onSelfHostedSourcegraphURLChange: (sourcegraphURL?: string) => void

    isActivated: boolean
    onToggleActivated: (value: boolean) => void

    isFullPage: boolean
    showPrivateRepositoryAlert?: boolean
    showSourcegraphCloudAlert?: boolean
    permissionAlert?: {
        name: string
        icon?: React.ComponentType<{ className?: string }>
    }
    requestPermissionsHandler?: React.MouseEventHandler
}

export const OptionsPage: React.FC<OptionsPageProps> = ({
    version,
    selfHostedSourcegraphURL,
    validateSourcegraphUrl,
    isActivated,
    onToggleActivated,
    isFullPage,
    showPrivateRepositoryAlert,
    showSourcegraphCloudAlert,
    permissionAlert,
    requestPermissionsHandler,
    onSelfHostedSourcegraphURLChange,
}) => {
    const [showAdvancedSettings, setShowAdvancedSettings] = useState(false)

    const toggleAdvancedSettings = useCallback(
        () => setShowAdvancedSettings(showAdvancedSettings => !showAdvancedSettings),
        []
    )

    const selfHostedSourcegraphUrlValidate = useCallback(
        (url: string) => {
            if (!url) {
                return of(undefined)
            }
            if (isCloudSourcegraphUrl(url)) {
                return of('Sourcegraph cloud is supported by default')
            }
            return validateSourcegraphUrl(url)
        },
        [validateSourcegraphUrl]
    )

    const handleFormSubmit = useCallback((event: React.FormEvent<HTMLFormElement>) => event.preventDefault(), [])

    return (
        <div className={classNames('options-page', isFullPage && 'options-page--full shadow')}>
            <section className="options-page__section">
                <div className="d-flex justify-content-between">
                    <a href={CLOUD_SOURCEGRAPH_URL} {...LINK_PROPS}>
                        <SourcegraphLogo className="options-page__logo" />
                    </a>
                    <div>
                        <Toggle
                            value={isActivated}
                            onToggle={onToggleActivated}
                            title={`Toggle to ${isActivated ? 'disable' : 'enable'} extension`}
                            aria-label="Toggle browser extension"
                        />
                    </div>
                </div>
                <div className="options-page__version">v{version}</div>
            </section>
            <InfoSection />
            <section className="options-page__section border-0">
                {/* eslint-disable-next-line react/forbid-elements */}
                <form onSubmit={handleFormSubmit} noValidate={true}>
                    <SourcegraphURLInput
                        id="sourcegraph-url"
                        label="Sourcegraph cloud"
                        editable={false}
                        className="mb-3"
                        validate={validateSourcegraphUrl}
                        initialValue={CLOUD_SOURCEGRAPH_URL}
                        description={
                            <>
                                Get code intel for millions of public repositories and your synced private repositories
                                on <a href={CLOUD_SOURCEGRAPH_URL}>{CLOUD_SOURCEGRAPH_URL.replace('https://', '')}</a>
                            </>
                        }
                    />
                    <SourcegraphURLInput
                        id="self-hosted-sourcegraph-url"
                        label="Sourcegraph self-hosted"
                        dataTestId="test-sourcegraph-url"
                        validate={selfHostedSourcegraphUrlValidate}
                        initialValue={selfHostedSourcegraphURL || ''}
                        onChange={onSelfHostedSourcegraphURLChange}
                        description="Enter the URL of your Sourcegraph instance to use the extension on a private instance."
                    />
                </form>
            </section>

            {permissionAlert && (
                <PermissionAlert {...permissionAlert} onClickGrantPermissions={requestPermissionsHandler} />
            )}

            {showSourcegraphCloudAlert && <SourcegraphCloudAlert />}

            {showPrivateRepositoryAlert && <PrivateRepositoryAlert />}
            <section className="options-page__section pt-2">
                <a
                    className="options-page__text-link display-flex align-items-center"
                    href="https://docs.sourcegraph.com/integration/browser_extension#privacy"
                    {...LINK_PROPS}
                >
                    <small>How do we keep your code private?</small>
                    <OpenInNewIcon className="options-page__open-new-icon mx-2" />
                </a>
                <p className="mb-0">
                    <button
                        type="button"
                        data-testid="test-show-advanced-settings"
                        className="options-page__text-link btn btn-link p-0"
                        onClick={toggleAdvancedSettings}
                    >
                        <small>{showAdvancedSettings ? 'Hide' : 'Show'} advanced settings</small>
                    </button>
                </p>
                {showAdvancedSettings && <OptionsPageAdvancedSettings />}
            </section>
        </div>
    )
}
