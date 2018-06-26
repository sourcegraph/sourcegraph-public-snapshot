import PencilIcon from '@sourcegraph/icons/lib/Pencil'
import marked from 'marked'
import * as React from 'react'
import { Link } from 'react-router-dom'
import * as GQL from '../backend/graphqlschema'
import { Markdown } from '../components/Markdown'
import { SourcegraphExtension } from '../schema/extension.schema'
import { parseJSON } from '../settings/configuration'
import { RegistryExtensionNoManifestAlert } from './RegistryExtensionManifestPage'

const EditManifestAlert: React.SFC<{
    extension: Pick<GQL.IRegistryExtension, 'viewerCanAdminister' | 'url'>
    text: string
    buttonLabel: string
    alertClass: 'alert-info' | 'alert-danger'
}> = ({ extension, text, buttonLabel, alertClass }) => (
    <div className={`alert ${alertClass}`}>
        {text}
        {extension.viewerCanAdminister && (
            <>
                <br />
                <Link className="mt-3 btn btn-primary" to={`${extension.url}/-/edit`}>
                    <PencilIcon className="icon-inline" /> {buttonLabel}
                </Link>
            </>
        )}
    </div>
)

export const RegistryExtensionDescription: React.SFC<{
    extension: Pick<GQL.IRegistryExtension, 'manifest' | 'viewerCanAdminister' | 'url'>
}> = ({ extension }) => {
    if (!extension.manifest || !extension.manifest.raw) {
        return <RegistryExtensionNoManifestAlert extension={extension} />
    }

    let manifest: SourcegraphExtension | null
    try {
        manifest = parseJSON(extension.manifest.raw)
    } catch (err) {
        return (
            <EditManifestAlert
                extension={extension}
                alertClass="alert-danger"
                text={`This extension's manifest is invalid: ${err && err.message ? err.message : 'JSON parse error'}`}
                buttonLabel="Edit manifest"
            />
        )
    }

    if (!manifest || !manifest.description) {
        return (
            <>
                {manifest && manifest.title && <h2>{manifest.title}</h2>}
                <EditManifestAlert
                    extension={extension}
                    alertClass="alert-info"
                    text={`This extension has no description.`}
                    buttonLabel="Add description"
                />
            </>
        )
    }

    try {
        const html = marked(manifest.description, { gfm: true, breaks: true, sanitize: true })
        return <Markdown dangerousInnerHTML={html} />
    } catch (err) {
        return (
            <EditManifestAlert
                extension={extension}
                alertClass="alert-danger"
                text="This extension's Markdown description is invalid."
                buttonLabel="Edit manifest"
            />
        )
    }
}
