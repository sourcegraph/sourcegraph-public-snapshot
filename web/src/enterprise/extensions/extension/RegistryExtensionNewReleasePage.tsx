import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import H from 'history'
import ErrorIcon from 'mdi-react/ErrorIcon'
import React, { useCallback, useEffect, useState } from 'react'
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
import { useLocalStorage } from '../../../util/useLocalStorage'

const publishExtension = (
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

        const [bundle, setBundle] = useState<string | ErrorLike>()

        useEffect(() => {
            if (extension.manifest && !isErrorLike(extension.manifest) && extension.manifest.url) {
                const extensionManifestUrl = extension.manifest.url
                ;(async () => {
                    const resp = await fetch(extensionManifestUrl)
                    setBundle(resp.status === 200 ? await resp.text() : '')
                })().catch(err => setBundle(asError(err)))
            } else {
                setBundle(undefined)
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
                    if (isErrorLike(bundle)) {
                        throw new Error('invalid bundle')
                    }
                    setUpdateOrError(await publishExtension({ extensionID: extension.id, manifest, bundle }))
                    onDidUpdateExtension()
                } catch (err) {
                    setUpdateOrError(asError(err))
                }
            },
            [extension.id, manifest, bundle, onDidUpdateExtension]
        )

        const [showEditor, setShowEditor] = useLocalStorage('RegistryExtensionNewReleasePage.showEditor', false)
        const onShowEditorClick = useCallback(() => setShowEditor(true), [setShowEditor])

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
                    <a href="https://github.com/sourcegraph/src-cli" target="_blank" rel="noopener noreferrer">
                        <code>src</code> CLI tool
                    </a>{' '}
                    to publish a new release:
                </p>
                <pre>
                    <code>$ src extensions publish</code>
                </pre>
                {showEditor ? (
                    <>
                        <hr className="my-4" />
                        <h2>Extension editor (experimental)</h2>
                        <p>
                            Edit or paste in an extension JSON manifest and JavaScript bundle. The JavaScript bundle
                            source must be self-contained; dependency bundling and TypeScript transpilation is not yet
                            supported.
                        </p>
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
                                        ) : isErrorLike(bundle) ? (
                                            <div className="alert alert-danger">{bundle.message}</div>
                                        ) : (
                                            <DynamicallyImportedMonacoSettingsEditor
                                                id="registry-extension-new-release-page__bundle"
                                                language="javascript"
                                                className="d-block"
                                                // Only 1 component can block navigation, and the
                                                // other editor does, so we don't.
                                                noBlockNavigationIfDirty={true}
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
                            <button
                                type="submit"
                                disabled={updateOrError === LOADING || isErrorLike(bundle)}
                                className="btn btn-primary"
                            >
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
                ) : (
                    <button type="button" className="btn btn-secondary" onClick={onShowEditorClick}>
                        Experimental: Use in-browser extension editor
                    </button>
                )}
            </div>
        )
    }
)
