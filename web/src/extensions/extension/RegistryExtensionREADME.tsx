import marked from 'marked'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { ConfiguredRegistryExtension } from '../../../../shared/src/extensions/extension'
import { isErrorLike } from '../../../../shared/src/util/errors'
import { ExtensionNoManifestAlert } from './RegistryExtensionManifestPage'

const PublishNewManifestAlert: React.FunctionComponent<{
    extension: ConfiguredRegistryExtension
    text: string
    buttonLabel: string
    alertClass: 'alert-info' | 'alert-danger'
}> = ({ extension, text, buttonLabel, alertClass }) => (
    <div className={`alert ${alertClass}`}>
        {text}
        {extension.registryExtension && extension.registryExtension.viewerCanAdminister && (
            <>
                <br />
                <Link className="mt-3 btn btn-primary" to={`${extension.registryExtension.url}/-/releases/new`}>
                    {buttonLabel}
                </Link>
            </>
        )}
    </div>
)

export const ExtensionREADME: React.FunctionComponent<{
    extension: ConfiguredRegistryExtension
}> = ({ extension }) => {
    if (!extension.rawManifest) {
        return <ExtensionNoManifestAlert extension={extension} />
    }

    const manifest = extension.manifest
    if (isErrorLike(manifest)) {
        return (
            <PublishNewManifestAlert
                extension={extension}
                alertClass="alert-danger"
                text={`This extension's manifest is invalid: ${
                    manifest && manifest.message ? manifest.message : 'JSON parse error'
                }`}
                buttonLabel="Fix manifest and publish new release"
            />
        )
    }

    if (!manifest || !manifest.readme) {
        return (
            <PublishNewManifestAlert
                extension={extension}
                alertClass="alert-info"
                text={`This extension has no README.`}
                buttonLabel="Add README and publish new release"
            />
        )
    }

    try {
        const html = marked(manifest.readme, { gfm: true, breaks: true, sanitize: true })
        return <Markdown dangerousInnerHTML={html} />
    } catch (err) {
        return (
            <PublishNewManifestAlert
                extension={extension}
                alertClass="alert-danger"
                text="This extension's Markdown README is invalid."
                buttonLabel="Fix README and publish new release"
            />
        )
    }
}
