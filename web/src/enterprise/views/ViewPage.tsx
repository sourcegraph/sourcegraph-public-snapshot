import React, { useMemo, useState, useEffect } from 'react'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { useView } from './useView'
import { ViewForm } from './forms/ViewForm'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../shared/src/util/markdown'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { PageTitle } from '../../components/PageTitle'
import H from 'history'
import { QueryInputInViewContent } from './QueryInputInViewContent'
import { MarkupContent } from 'sourcegraph'
import { CaseSensitivityProps, PatternTypeProps } from '../../search'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'

interface Props
    extends ExtensionsControllerProps<'services'>,
        SettingsCascadeProps,
        PatternTypeProps,
        CaseSensitivityProps {
    viewID: string
    extraPath: string

    location: H.Location
    history: H.History
}

/**
 * A page that displays a single view (contributed by an extension).
 */
export const ViewPage: React.FunctionComponent<Props> = ({
    viewID,
    extraPath,
    location,
    extensionsController,
    ...props
}) => {
    const queryParams = useMemo<{ [key: string]: string }>(
        () => ({ ...Object.fromEntries(new URLSearchParams(location.search).entries()), extraPath }),
        [extraPath, location.search]
    )
    const data = useView(
        viewID,
        queryParams,
        useMemo(() => extensionsController.services.contribution.getContributions(), [
            extensionsController.services.contribution,
        ]),
        extensionsController.services.view
    )

    // Wait for extensions to load for up to 5 seconds before showing "not found".
    const [waited, setWaited] = useState(false)
    useEffect(() => {
        setTimeout(() => setWaited(true), 5000)
    }, [])

    if (data === undefined || (!waited && data === null)) {
        return null
    }
    if (data === null) {
        return (
            <div className="alert alert-danger">
                View not found: <code>{viewID}</code>
            </div>
        )
    }

    const { contribution, form, view } = data

    const title = view?.title || contribution.title

    return (
        <div>
            <PageTitle title={title || 'View'} />
            {title && <h1>{title}</h1>}
            {form === undefined ? null : form === null ? (
                <div className="alert alert-danger">
                    View form not found: <code>{contribution.form}</code>
                </div>
            ) : (
                <ViewForm form={form} extensionsController={extensionsController} />
            )}
            {view?.content.map((content, i) =>
                isMarkupContent(content) ? (
                    <section key={i} className="mt-3">
                        {content.kind === MarkupKind.Markdown ? (
                            <Markdown dangerousInnerHTML={renderMarkdown(content.value)} />
                        ) : (
                            content.value
                        )}
                    </section>
                ) : content.component === 'QueryInput' ? (
                    <QueryInputInViewContent
                        {...props}
                        location={location}
                        implicitQueryPrefix={
                            typeof content.props.implicitQueryPrefix === 'string'
                                ? content.props.implicitQueryPrefix
                                : ''
                        }
                    />
                ) : null
            )}
        </div>
    )
}

function isMarkupContent(v: unknown): v is MarkupContent {
    return (
        typeof v === 'object' &&
        v !== null &&
        'kind' in v &&
        typeof (v as any).kind === 'string' &&
        typeof (v as any).value === 'string'
    )
}
