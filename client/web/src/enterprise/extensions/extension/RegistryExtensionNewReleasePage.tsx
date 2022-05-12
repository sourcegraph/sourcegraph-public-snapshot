import React, { useCallback, useState } from 'react'

import * as H from 'history'
import AlertCircleIcon from 'mdi-react/AlertCircleIcon'
import CheckCircleIcon from 'mdi-react/CheckCircleIcon'
import { of, Observable, concat, from } from 'rxjs'
import { fromFetch } from 'rxjs/fetch'
import { map, catchError, tap, concatMap } from 'rxjs/operators'

import { ErrorAlert } from '@sourcegraph/branded/src/components/alerts'
import { Form } from '@sourcegraph/branded/src/components/Form'
import { asError, isErrorLike } from '@sourcegraph/common'
import { gql, dataOrThrowErrors } from '@sourcegraph/http-client'
import { ConfiguredRegistryExtension } from '@sourcegraph/shared/src/extensions/extension'
import { ExtensionManifest } from '@sourcegraph/shared/src/extensions/extensionManifest'
import * as GQL from '@sourcegraph/shared/src/schema'
import extensionSchemaJSON from '@sourcegraph/shared/src/schema/extension.schema.json'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { ThemeProps } from '@sourcegraph/shared/src/theme'
import {
    Button,
    LoadingSpinner,
    useLocalStorage,
    useEventObservable,
    Link,
    Icon,
    Typography,
} from '@sourcegraph/wildcard'

import { AuthenticatedUser } from '../../../auth'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { mutateGraphQL } from '../../../backend/graphql'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'

const publishExtension = (
    args: Pick<GQL.IPublishExtensionOnExtensionRegistryMutationArguments, 'extensionID' | 'manifest' | 'bundle'>
): Promise<GQL.IExtensionRegistryPublishExtensionResult> =>
    mutateGraphQL(
        gql`
            mutation PublishRegistryExtension($extensionID: String!, $manifest: String!, $bundle: String!) {
                extensionRegistry {
                    publishExtension(extensionID: $extensionID, manifest: $manifest, bundle: $bundle) {
                        extension {
                            url
                        }
                    }
                }
            }
        `,
        args
    )
        .pipe(
            map(dataOrThrowErrors),
            map(data => data.extensionRegistry.publishExtension)
        )
        .toPromise()

interface Props extends ThemeProps, TelemetryProps {
    /** The extension that is the subject of the page. */
    extension: ConfiguredRegistryExtension<GQL.IRegistryExtension>

    authenticatedUser: AuthenticatedUser
    history: H.History
}

const DEFAULT_MANIFEST: Pick<ExtensionManifest, 'activationEvents'> = {
    activationEvents: ['*'],
}

const LOADING = 'loading' as const

const DEFAULT_SOURCE = `const sourcegraph = require('sourcegraph')

function activate(context) {
    sourcegraph.app.activeWindow.showNotification(
        'Hello World!',
        sourcegraph.NotificationType.Success
    )
}

module.exports = { activate }
`

/** A page for publishing a new release of an extension to the extension registry. */
export const RegistryExtensionNewReleasePage = withAuthenticatedUser<Props>(
    ({ extension, isLightTheme, telemetryService, history }) => {
        // Omit the `url` field from the extension so that it gets set to the URL of the bundle we're uploading.
        const manifestWithoutUrl = extension.rawManifest ? JSON.parse(extension.rawManifest) : { ...DEFAULT_MANIFEST }
        delete manifestWithoutUrl.url
        const [manifest, setManifest] = useState(JSON.stringify(manifestWithoutUrl, null, 2))

        const [onChangeBundle, bundleOrError] = useEventObservable(
            useCallback(
                (bundleChanges: Observable<string>) =>
                    concat(
                        isErrorLike(extension.manifest) || !extension.manifest?.url
                            ? of(DEFAULT_SOURCE)
                            : fromFetch(extension.manifest.url, { selector: response => response.text() }).pipe(
                                  catchError(error => [asError(error)])
                              ),
                        bundleChanges
                    ),
                [extension.manifest]
            )
        )

        const [onSubmit, updateOrError] = useEventObservable(
            useCallback(
                (submits: Observable<React.FormEvent>) =>
                    submits.pipe(
                        tap(event => event.preventDefault()),
                        concatMap(() => {
                            if (isErrorLike(bundleOrError)) {
                                throw new Error('Invalid bundle')
                            }
                            return concat(
                                [LOADING],
                                from(
                                    publishExtension({ extensionID: extension.id, manifest, bundle: bundleOrError })
                                ).pipe(catchError(error => [asError(error)]))
                            )
                        })
                    ),
                [bundleOrError, extension.id, manifest]
            )
        )

        const [showEditor, setShowEditor] = useLocalStorage('RegistryExtensionNewReleasePage.showEditor', false)
        const onShowEditorClick = useCallback(() => setShowEditor(true), [setShowEditor])

        return !extension.registryExtension || !extension.registryExtension.viewerCanAdminister ? (
            <HeroPage
                icon={AlertCircleIcon}
                title="Unauthorized"
                subtitle="You are not authorized to adminster this extension."
            />
        ) : (
            <div className="registry-extension-new-release-page">
                <PageTitle title="Publish new release" />
                <Typography.H2>Publish new release</Typography.H2>
                <p>
                    Use the{' '}
                    <Link to="https://github.com/sourcegraph/src-cli" target="_blank" rel="noopener noreferrer">
                        <code>src</code> CLI tool
                    </Link>{' '}
                    to publish a new release:
                </p>
                <pre>
                    <code>$ src extensions publish</code>
                </pre>
                {showEditor ? (
                    <>
                        <hr className="my-4" />
                        <Typography.H2>Extension editor (experimental)</Typography.H2>
                        <p>
                            Edit or paste in an extension JSON manifest and JavaScript bundle. The JavaScript bundle
                            source must be self-contained; dependency bundling and TypeScript transpilation is not yet
                            supported.
                        </p>
                        <Form onSubmit={onSubmit} className="mb-3">
                            <div className="row">
                                <div className="col-lg-6">
                                    <div className="form-group">
                                        <label htmlFor="registry-extension-new-release-page__manifest">
                                            <Typography.H3>Manifest</Typography.H3>
                                        </label>
                                        <DynamicallyImportedMonacoSettingsEditor
                                            id="registry-extension-new-release-page__manifest"
                                            className="d-block"
                                            value={manifest}
                                            onChange={setManifest}
                                            jsonSchema={extensionSchemaJSON}
                                            readOnly={updateOrError === LOADING}
                                            isLightTheme={isLightTheme}
                                            history={history}
                                            telemetryService={telemetryService}
                                        />
                                    </div>
                                </div>
                                <div className="col-lg-6">
                                    <div className="form-group">
                                        <label htmlFor="registry-extension-new-release-page__bundle">
                                            <Typography.H3>Source</Typography.H3>
                                        </label>
                                        {bundleOrError === undefined ? (
                                            <div>
                                                <LoadingSpinner />
                                            </div>
                                        ) : isErrorLike(bundleOrError) ? (
                                            <ErrorAlert error={bundleOrError} />
                                        ) : (
                                            <DynamicallyImportedMonacoSettingsEditor
                                                id="registry-extension-new-release-page__bundle"
                                                language="javascript"
                                                className="d-block"
                                                // Only 1 component can block navigation, and the
                                                // other editor does, so we don't.
                                                blockNavigationIfDirty={false}
                                                value={bundleOrError}
                                                onChange={onChangeBundle}
                                                readOnly={updateOrError === LOADING}
                                                isLightTheme={isLightTheme}
                                                history={history}
                                                telemetryService={telemetryService}
                                            />
                                        )}
                                    </div>
                                </div>
                            </div>
                            <div aria-live="polite" className="d-flex align-items-center">
                                <Button
                                    type="submit"
                                    disabled={updateOrError === LOADING || isErrorLike(bundleOrError)}
                                    className="mr-2"
                                    variant="primary"
                                >
                                    Publish
                                </Button>{' '}
                                {updateOrError &&
                                    !isErrorLike(updateOrError) &&
                                    (updateOrError === LOADING ? (
                                        <LoadingSpinner />
                                    ) : (
                                        <span className="text-success">
                                            <Icon as={CheckCircleIcon} /> Published release successfully.
                                        </span>
                                    ))}
                            </div>
                            {isErrorLike(updateOrError) && <ErrorAlert error={updateOrError} />}
                        </Form>
                    </>
                ) : (
                    <Button onClick={onShowEditorClick} variant="secondary">
                        Experimental: Use in-browser extension editor
                    </Button>
                )}
            </div>
        )
    }
)
