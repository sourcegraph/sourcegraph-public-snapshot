import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react'

import { Combobox, ComboboxInput, ComboboxOption, ComboboxPopover, ComboboxList } from '@reach/combobox'
import classNames from 'classnames'
import BlockHelperIcon from 'mdi-react/BlockHelperIcon'
import BookOpenPageVariantIcon from 'mdi-react/BookOpenPageVariantIcon'
import CheckCircleOutlineIcon from 'mdi-react/CheckCircleOutlineIcon'
import EarthIcon from 'mdi-react/EarthIcon'
import LockIcon from 'mdi-react/LockIcon'
import OpenInNewIcon from 'mdi-react/OpenInNewIcon'
import { Observable } from 'rxjs'

import { LoaderInput } from '@sourcegraph/branded/src/components/LoaderInput'
import { SourcegraphLogo } from '@sourcegraph/branded/src/components/SourcegraphLogo'
import { Toggle } from '@sourcegraph/branded/src/components/Toggle'
import { IUser } from '@sourcegraph/shared/src/schema'
import { createURLWithUTM } from '@sourcegraph/shared/src/tracking/utm'
import { useInputValidation, deriveInputClassName } from '@sourcegraph/shared/src/util/useInputValidation'
import { Button, Link, Icon, Typography } from '@sourcegraph/wildcard'

import { getPlatformName, isDefaultSourcegraphUrl } from '../../shared/util/context'

import { OptionsPageContainer } from './components/OptionsPageContainer'
import { OptionsPageAdvancedSettings } from './OptionsPageAdvancedSettings'

import styles from './OptionsPage.module.scss'

import '@reach/combobox/styles.css'

export interface OptionsPageProps {
    version: string

    // Sourcegraph URL
    sourcegraphUrl: string
    validateSourcegraphUrl: (url: string) => Observable<string | undefined>
    onChangeSourcegraphUrl: (url: string) => void

    // Suggested Sourcegraph URLs
    suggestedSourcegraphUrls: string[]

    // Option flags
    optionFlags: { key: string; label: string; value: boolean }[]
    onChangeOptionFlag: (key: string, value: boolean) => void

    isActivated: boolean
    onToggleActivated: (value: boolean) => void

    initialShowAdvancedSettings?: boolean
    isFullPage: boolean
    showSourcegraphCloudAlert?: boolean
    permissionAlert?: { name: string; icon?: React.ComponentType<{ className?: string }> }
    requestPermissionsHandler?: React.MouseEventHandler

    hasRepoSyncError?: boolean
    currentUser?: Pick<IUser, 'settingsURL' | 'siteAdmin'>
}

// "Error code" constants for Sourcegraph URL validation
export const URL_FETCH_ERROR = 'URL_FETCH_ERROR'
export const URL_AUTH_ERROR = 'URL_AUTH_ERROR'

const NEW_TAB_LINK_PROPS: Pick<React.AnchorHTMLAttributes<HTMLAnchorElement>, 'rel' | 'target'> = {
    target: '_blank',
    rel: 'noopener noreferrer',
}

export const OptionsPage: React.FunctionComponent<React.PropsWithChildren<OptionsPageProps>> = ({
    version,
    sourcegraphUrl,
    validateSourcegraphUrl,
    isActivated,
    onToggleActivated,
    initialShowAdvancedSettings = false,
    isFullPage,
    showSourcegraphCloudAlert,
    permissionAlert,
    requestPermissionsHandler,
    optionFlags,
    onChangeOptionFlag,
    onChangeSourcegraphUrl,
    suggestedSourcegraphUrls,
    hasRepoSyncError,
    currentUser,
}) => {
    const [showAdvancedSettings, setShowAdvancedSettings] = useState(initialShowAdvancedSettings)

    const toggleAdvancedSettings = useCallback(
        () => setShowAdvancedSettings(showAdvancedSettings => !showAdvancedSettings),
        []
    )

    return (
        <OptionsPageContainer className="shadow" isFullPage={isFullPage}>
            <section className={classNames(styles.section, 'pb-2')}>
                <div className="d-flex justify-content-between">
                    <SourcegraphLogo className={styles.logo} />
                    <div>
                        <Toggle
                            value={isActivated}
                            onToggle={onToggleActivated}
                            title={`Toggle to ${isActivated ? 'disable' : 'enable'} extension`}
                            aria-label="Toggle browser extension"
                        />
                    </div>
                </div>
                <div className={styles.version}>v{version}</div>
            </section>
            <section className={styles.section}>
                Get code intelligence tooltips while browsing and reviewing code on your code host.{' '}
                <Link to="https://docs.sourcegraph.com/integration/browser_extension#features" {...NEW_TAB_LINK_PROPS}>
                    Learn more
                </Link>{' '}
                about the extension and compatible code hosts.
            </section>
            <section className={classNames('border-0', styles.section)}>
                <SourcegraphURLForm
                    value={sourcegraphUrl}
                    suggestions={suggestedSourcegraphUrls}
                    onChange={onChangeSourcegraphUrl}
                    validate={validateSourcegraphUrl}
                />
                <p className="mt-2 mb-0">
                    <small>Enter the URL of your Sourcegraph instance to use the extension on private code.</small>
                </p>
            </section>

            {permissionAlert && (
                <PermissionAlert {...permissionAlert} onClickGrantPermissions={requestPermissionsHandler} />
            )}

            {showSourcegraphCloudAlert && <SourcegraphCloudAlert />}

            {hasRepoSyncError && currentUser && (
                <RepoSyncErrorAlert sourcegraphUrl={sourcegraphUrl} currentUser={currentUser} />
            )}

            <section className={styles.section}>
                <Link
                    to="https://docs.sourcegraph.com/integration/browser_extension#privacy"
                    {...NEW_TAB_LINK_PROPS}
                    className="d-block mb-1"
                >
                    <small>How do we keep your code private?</small> <OpenInNewIcon size="0.75rem" className="ml-2" />
                </Link>
                <p className="mb-0">
                    <Button
                        className="p-0 shadow-none font-weight-normal test-toggle-advanced-settings-button"
                        onClick={toggleAdvancedSettings}
                        variant="link"
                        size="sm"
                    >
                        {showAdvancedSettings ? 'Hide' : 'Show'} advanced settings
                    </Button>
                </p>
                {showAdvancedSettings && (
                    <OptionsPageAdvancedSettings optionFlags={optionFlags} onChangeOptionFlag={onChangeOptionFlag} />
                )}
            </section>
            <section className="d-flex">
                <div className={styles.splitSectionPart}>
                    <Link to="https://sourcegraph.com/search" {...NEW_TAB_LINK_PROPS}>
                        <Icon role="img" className="mr-2" as={EarthIcon} aria-hidden={true} />
                        Sourcegraph Cloud
                    </Link>
                </div>
                <div className={styles.splitSectionPart}>
                    <Link to="https://docs.sourcegraph.com" {...NEW_TAB_LINK_PROPS}>
                        <Icon role="img" className="mr-2" as={BookOpenPageVariantIcon} aria-hidden={true} />
                        Documentation
                    </Link>
                </div>
            </section>
        </OptionsPageContainer>
    )
}

interface PermissionAlertProps {
    icon?: React.ComponentType<React.PropsWithChildren<{ className?: string }>>
    name: string
    onClickGrantPermissions?: React.MouseEventHandler
}

const PermissionAlert: React.FunctionComponent<React.PropsWithChildren<PermissionAlertProps>> = ({
    name,
    icon: AlertIcon,
    onClickGrantPermissions,
}) => (
    <section className={classNames('bg-2', styles.section)}>
        <Typography.H4>
            {AlertIcon && <Icon role="img" className="mr-2" as={AlertIcon} aria-hidden={true} />} <span>{name}</span>
        </Typography.H4>
        <p className={styles.permissionText}>
            <strong>Grant permissions</strong> to use the Sourcegraph extension on {name}.
        </p>
        <Button onClick={onClickGrantPermissions} variant="primary" size="sm">
            <small>Grant permissions</small>
        </Button>
    </section>
)

const RepoSyncErrorAlert: React.FunctionComponent<
    React.PropsWithChildren<{
        sourcegraphUrl: OptionsPageProps['sourcegraphUrl']
        currentUser: NonNullable<OptionsPageProps['currentUser']>
    }>
> = ({ sourcegraphUrl, currentUser }) => {
    const isDefaultURL = isDefaultSourcegraphUrl(sourcegraphUrl)

    if (isDefaultURL && !currentUser.settingsURL) {
        return null
    }

    return (
        <section className={classNames('bg-2', styles.section)}>
            <Typography.H4>
                <Icon role="img" aria-hidden={true} className="mr-2" as={isDefaultURL ? LockIcon : BlockHelperIcon} />
                {isDefaultURL ? 'Private repository' : 'Repository not found'}
            </Typography.H4>
            <p className="mb-0">
                {isDefaultURL ? (
                    <>
                        <Link
                            to={
                                createURLWithUTM(
                                    new URL(`${currentUser.settingsURL!}/repositories/manage`, sourcegraphUrl),
                                    {
                                        utm_source: getPlatformName(),
                                        utm_campaign: 'sync-private-repo-with-cloud',
                                    }
                                ).href
                            }
                            {...NEW_TAB_LINK_PROPS}
                            className={styles.link}
                        >
                            Add your repository to Sourcegraph
                        </Link>{' '}
                        to use this extension for private repositories.
                    </>
                ) : currentUser.siteAdmin ? (
                    <>
                        <Link
                            to={
                                createURLWithUTM(new URL('admin/repo/add', 'https://docs.sourcegraph.com/'), {
                                    utm_source: getPlatformName(),
                                    utm_campaign: 'add-repo-to-instance',
                                }).href
                            }
                            {...NEW_TAB_LINK_PROPS}
                            className={styles.link}
                        >
                            Add your repository to Sourcegraph
                        </Link>{' '}
                        to use this extension.
                    </>
                ) : (
                    <>Contact your admin to add this repository to Sourcegraph.</>
                )}
            </p>
        </section>
    )
}

const SourcegraphCloudAlert: React.FunctionComponent<React.PropsWithChildren<unknown>> = () => (
    <section className={classNames('bg-2', styles.section)}>
        <Typography.H4>
            <Icon role="img" aria-hidden={true} className="mr-2" as={CheckCircleOutlineIcon} />
            You're on Sourcegraph Cloud
        </Typography.H4>
        <p>Naturally, the browser extension is not necessary to browse public code on sourcegraph.com.</p>
    </section>
)

function preventDefault(event: React.FormEvent<HTMLFormElement>): void {
    event.preventDefault()
}

interface SourcegraphURLFormProps {
    value: OptionsPageProps['sourcegraphUrl']
    validate: OptionsPageProps['validateSourcegraphUrl']
    onChange: OptionsPageProps['onChangeSourcegraphUrl']
    suggestions: OptionsPageProps['sourcegraphUrl'][]
}

export const SourcegraphURLForm: React.FunctionComponent<React.PropsWithChildren<SourcegraphURLFormProps>> = ({
    value,
    validate,
    suggestions,
    onChange,
}) => {
    const urlInputReference = useRef<HTMLInputElement | null>(null)

    const [urlState, nextUrlFieldChange, nextUrlInputElement] = useInputValidation(
        useMemo(
            () => ({
                initialValue: value,
                synchronousValidators: [],
                asynchronousValidators: [validate],
            }),
            [value, validate]
        )
    )

    const urlInputElements = useCallback(
        (urlInputElement: HTMLInputElement | null) => {
            urlInputReference.current = urlInputElement
            nextUrlInputElement(urlInputElement)
        },
        [nextUrlInputElement]
    )

    /**
     * BEGIN: Workaround for reach/combobox undesirably expanded
     *
     * @see https://github.com/reach/reach-ui/issues/755
     */
    const [hasInteracted, setHasInteracted] = useState(false)
    const onFocus = useCallback(() => {
        if (!hasInteracted) {
            setHasInteracted(true)
        }
    }, [hasInteracted])
    /**
     * END: Workaround for reach/combobox undesirably expanded
     */

    useEffect(() => {
        if (urlState.kind === 'VALID') {
            onChange(urlState.value)
        }
    }, [onChange, urlState])

    return (
        // eslint-disable-next-line react/forbid-elements
        <form onSubmit={preventDefault} noValidate={true}>
            <label htmlFor="sourcegraph-url">Sourcegraph URL</label>
            <Combobox openOnFocus={true} onSelect={nextUrlFieldChange}>
                <LoaderInput loading={urlState.kind === 'LOADING'} className={deriveInputClassName(urlState)}>
                    <ComboboxInput
                        type="url"
                        required={true}
                        spellCheck={false}
                        autoComplete="off"
                        autocomplete={false}
                        pattern="^https://.*"
                        placeholder="https://"
                        onFocus={onFocus}
                        id="sourcegraph-url"
                        ref={urlInputElements}
                        value={urlState.value}
                        onChange={nextUrlFieldChange}
                        className={classNames('form-control', 'test-sourcegraph-url', deriveInputClassName(urlState))}
                    />
                </LoaderInput>

                {suggestions.length > 1 && hasInteracted && (
                    <ComboboxPopover className={styles.popover}>
                        <ComboboxList>
                            {suggestions.map(suggestion => (
                                <ComboboxOption key={suggestion} value={suggestion} />
                            ))}
                        </ComboboxList>
                    </ComboboxPopover>
                )}
            </Combobox>
            <div className="mt-2">
                {urlState.kind === 'LOADING' ? (
                    <small className="d-block text-muted">Checking...</small>
                ) : urlState.kind === 'INVALID' ? (
                    <small className="d-block invalid-feedback">
                        {urlState.reason === URL_FETCH_ERROR ? (
                            'Incorrect Sourcegraph instance address'
                        ) : urlState.reason === URL_AUTH_ERROR ? (
                            <>
                                Authentication to Sourcegraph failed.{' '}
                                <Link to={urlState.value} {...NEW_TAB_LINK_PROPS}>
                                    Sign in to your instance
                                </Link>{' '}
                                to continue
                            </>
                        ) : urlInputReference.current?.validity.typeMismatch ? (
                            'Please enter a valid URL, including the protocol prefix (e.g. https://sourcegraph.example.com).'
                        ) : urlInputReference.current?.validity.patternMismatch ? (
                            'The browser extension can only work over HTTPS in modern browsers.'
                        ) : (
                            urlState.reason
                        )}
                    </small>
                ) : (
                    <small className="d-block valid-feedback test-valid-sourcegraph-url-feedback">Looks good!</small>
                )}
            </div>
        </form>
    )
}
