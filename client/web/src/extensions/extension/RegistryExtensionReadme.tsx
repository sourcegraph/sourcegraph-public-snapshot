import * as React from 'react'

import { isErrorLike, renderMarkdown } from '@sourcegraph/common'
import { Markdown } from '@sourcegraph/shared/src/components/Markdown'
import { ConfiguredRegistryExtension } from '@sourcegraph/shared/src/extensions/extension'
import { Button, Link, Alert } from '@sourcegraph/wildcard'

import { ExtensionNoManifestAlert } from './RegistryExtensionManifestPage'

const PublishNewManifestAlert: React.FunctionComponent<
    React.PropsWithChildren<{
        extension: ConfiguredRegistryExtension
        text: string
        buttonLabel: string
        alertVariant: 'info' | 'danger'
    }>
> = ({ extension, text, buttonLabel, alertVariant }) => (
    <Alert variant={alertVariant}>
        {text}
        {extension.registryExtension?.viewerCanAdminister && (
            <>
                <br />
                <Button
                    className="mt-3"
                    to={`${extension.registryExtension.url}/-/releases/new`}
                    variant="primary"
                    as={Link}
                >
                    {buttonLabel}
                </Button>
            </>
        )}
    </Alert>
)

export const ExtensionReadme: React.FunctionComponent<
    React.PropsWithChildren<{
        extension: ConfiguredRegistryExtension
    }>
> = ({ extension }) => {
    if (!extension.rawManifest) {
        return <ExtensionNoManifestAlert extension={extension} />
    }

    const manifest = extension.manifest
    if (isErrorLike(manifest)) {
        return (
            <PublishNewManifestAlert
                extension={extension}
                alertVariant="danger"
                text={`This extension's manifest is invalid: ${
                    manifest?.message ? manifest.message : 'JSON parse error'
                }`}
                buttonLabel="Fix manifest and publish new release"
            />
        )
    }

    if (!manifest || !manifest.readme) {
        return (
            <PublishNewManifestAlert
                extension={extension}
                alertVariant="info"
                text="This extension has no README."
                buttonLabel="Add README and publish new release"
            />
        )
    }

    try {
        const html = renderMarkdown(manifest.readme)
        return <Markdown testId="registry-extension-overview" dangerousInnerHTML={html} />
    } catch {
        return (
            <PublishNewManifestAlert
                extension={extension}
                alertVariant="danger"
                text="This extension's Markdown README is invalid."
                buttonLabel="Fix README and publish new release"
            />
        )
    }
}
