import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import ErrorIcon from 'mdi-react/ErrorIcon'
import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { map } from 'rxjs/operators'
import { ConfiguredRegistryExtension } from '../../../../../shared/src/extensions/extension'
import { ExtensionManifest } from '../../../../../shared/src/extensions/extensionManifest'
import { gql } from '../../../../../shared/src/graphql/graphql'
import * as GQL from '../../../../../shared/src/graphql/schema'
import extensionSchemaJSON from '../../../../../shared/src/schema/extension.schema.json'
import { asError, createAggregateError, ErrorLike, isErrorLike } from '../../../../../shared/src/util/errors'
import { withAuthenticatedUser } from '../../../auth/withAuthenticatedUser'
import { mutateGraphQL } from '../../../backend/graphql'
import { Form } from '../../../components/Form'
import { HeroPage } from '../../../components/HeroPage'
import { PageTitle } from '../../../components/PageTitle'
import { DynamicallyImportedMonacoSettingsEditor } from '../../../settings/DynamicallyImportedMonacoSettingsEditor'

const publishExtension = async (
    args: Pick<GQL.IPublishExtensionOnExtensionRegistryMutationArguments, 'extensionID' | 'manifest' | 'bundle'>
): Promise<GQL.IExtensionRegistryCreateExtensionResult> =>
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
            map(({ data, errors }) => {
                if (
                    !data ||
                    !data.extensionRegistry ||
                    !data.extensionRegistry.publishExtension ||
                    (errors && errors.length > 0)
                ) {
                    throw createAggregateError(errors)
                }
                return data.extensionRegistry.publishExtension
            })
        )
        .toPromise()

interface Props {
    /** The extension that is the subject of the page. */
    extension: ConfiguredRegistryExtension<GQL.IRegistryExtension>

    onDidUpdateExtension: () => void

    authenticatedUser: GQL.IUser
    isLightTheme: boolean
    history: H.History
}

const EXP_EXTENSION_EDITOR = localStorage.getItem('expExtensionEditor') !== null

const DEFAULT_MANIFEST: Pick<ExtensionManifest, 'activationEvents'> = {
    activationEvents: ['*'],
}

const LOADING: 'loading' = 'loading'

/** A page for publishing a new release of an extension to the extension registry. */
// tslint:disable: react-hooks-nesting
export const RegistryExtensionNewReleasePage = withAuthenticatedUser<Props>(
    ({ extension, onDidUpdateExtension, isLightTheme, history }) => {
        const [updateOrError, setUpdateOrError] = useState<
            null | typeof LOADING | GQL.IExtensionRegistryCreateExtensionResult | ErrorLike
        >(null)

        // Omit the `url` field from the extension so that it gets set to the URL of the bundle
        // we're uploading.
        const manifestWithoutUrl = extension.rawManifest ? JSON.parse(extension.rawManifest) : { ...DEFAULT_MANIFEST }
        delete manifestWithoutUrl.url
        const [manifest, setManifest] = useState(JSON.stringify(manifestWithoutUrl, null, 2))

        const [bundle, setBundle] = useState<string>()

        // tslint:disable-next-line: no-floating-promises
        useMemo(async () => {
            if (extension.manifest && !isErrorLike(extension.manifest) && extension.manifest.url) {
                try {
                    const resp = await fetch(extension.manifest.url)
                    setBundle(resp.status === 200 ? await resp.text() : '')
                } catch (err) {
                    setBundle('')
                }
            } else {
                setBundle('')
            }
        }, [extension])

        useEffect(() => {
            setUpdateOrError(null) // reset
        }, [manifest, bundle])

        const onSubmit = useCallback<React.FormEventHandler>(
            async e => {
                e.preventDefault()
                setUpdateOrError(LOADING)
                try {
                    setUpdateOrError(await publishExtension({ extensionID: extension.id, manifest, bundle }))
                    onDidUpdateExtension()
                } catch (err) {
                    setUpdateOrError(asError(err))
                }
            },
            [manifest, bundle]
        )

        return !extension.registryExtension || !extension.registryExtension.viewerCanAdminister ? (
            <HeroPage
                icon={ErrorIcon}
                title="Unauthorized"
                subtitle="You are not authorized to adminster this extension."
            />
        ) : (
            <div className="registry-extension-new-release-page">
                <PageTitle title="Publish new release" />
                <h2>Publish new release</h2>
                <p>
                    Use the{' '}
                    <a href="https://github.com/sourcegraph/src-cli" target="_blank">
                        <code>src</code> CLI tool
                    </a>{' '}
                    to publish a new release:
                </p>
                <pre>
                    <code>$ src extensions publish</code>
                </pre>
                {EXP_EXTENSION_EDITOR && (
                    <>
                        <hr className="my-4" />
                        <h2>Extension editor</h2>
                        <Form onSubmit={onSubmit} className="mb-3">
                            <div className="row">
                                <div className="col-lg-6">
                                    <div className="form-group">
                                        <label htmlFor="registry-extension-new-release-page__manifest">Manifest</label>
                                        <DynamicallyImportedMonacoSettingsEditor
                                            id="registry-extension-new-release-page__manifest"
                                            className="d-block"
                                            value={manifest}
                                            onChange={setManifest}
                                            jsonSchema={extensionSchemaJSON}
                                            readOnly={updateOrError === LOADING}
                                            isLightTheme={isLightTheme}
                                            history={history}
                                        />
                                    </div>
                                </div>
                                <div className="col-lg-6">
                                    <div className="form-group">
                                        <label htmlFor="registry-extension-new-release-page__bundle">Source</label>
                                        {bundle === undefined ? (
                                            <div>
                                                <LoadingSpinner className="icon-inline" />
                                            </div>
                                        ) : (
                                            <DynamicallyImportedMonacoSettingsEditor
                                                id="registry-extension-new-release-page__bundle"
                                                language="javascript"
                                                className="d-block"
                                                value={bundle}
                                                onChange={setBundle}
                                                readOnly={updateOrError === LOADING}
                                                isLightTheme={isLightTheme}
                                                history={history}
                                            />
                                        )}
                                    </div>
                                </div>
                            </div>
                            <button type="submit" disabled={updateOrError === LOADING} className="btn btn-primary">
                                {updateOrError === LOADING ? <LoadingSpinner className="icon-inline" /> : 'Publish'}
                            </button>
                        </Form>
                        {updateOrError &&
                            (isErrorLike(updateOrError) ? (
                                <div className="alert alert-danger">{updateOrError.message}</div>
                            ) : (
                                updateOrError !== LOADING && (
                                    <div className="alert alert-success">Published release successfully.</div>
                                )
                            ))}
                    </>
                )}
            </div>
        )
    }
)
