import React, { useMemo } from 'react'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { useView } from './useView'
import { Markdown } from '../../../../shared/src/components/Markdown'
import { renderMarkdown } from '../../../../shared/src/util/markdown'
import { MarkupKind } from '@sourcegraph/extension-api-classes'
import { PageTitle } from '../../components/PageTitle'
import H from 'history'
import { QueryInputInViewContent } from './QueryInputInViewContent'
import { MarkupContent } from 'sourcegraph'
import { CaseSensitivityProps, PatternTypeProps } from '../../search'
import { SettingsCascadeProps } from '../../../../shared/src/settings/settings'
import { ContributableViewContainer } from '../../../../shared/src/api/protocol'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { hasProperty } from '../../../../shared/src/util/types'

interface Props
    extends ExtensionsControllerProps<'services'>,
        SettingsCascadeProps,
        PatternTypeProps,
        CaseSensitivityProps {
    viewID: string
    extraPath: string

    location: H.Location
    history: H.History

    /** For mocking in tests. */
    _useView?: typeof useView
}

/**
 * A page that displays a single view (contributed by an extension) as a standalone page.
 */
export const ViewPage: React.FunctionComponent<Props> = ({
    viewID,
    extraPath,
    location,
    extensionsController,
    _useView = useView,
    ...props
}) => {
    const queryParams = useMemo<Record<string, string>>(
        () => ({ ...Object.fromEntries(new URLSearchParams(location.search).entries()), extraPath }),
        [extraPath, location.search]
    )
    const view = _useView(
        viewID,
        ContributableViewContainer.Global,
        queryParams,
        useMemo(() => extensionsController.services.contribution.getContributions(), [
            extensionsController.services.contribution,
        ]),
        extensionsController.services.view
    )

    if (view === undefined) {
        return <LoadingSpinner className="icon-inline" />
    }

    if (view === null) {
        return (
            <div className="alert alert-danger">
                View not found: <code>{viewID}</code>
            </div>
        )
    }

    return (
        <div>
            <PageTitle title={view.title || 'View'} />
            {view.title && <h1>{view.title}</h1>}
            {view.content.map((content, i) =>
                isMarkupContent(content) ? (
                    <section key={i} className="mt-3">
                        {content.kind === MarkupKind.Markdown || !content.kind ? (
                            <Markdown dangerousInnerHTML={renderMarkdown(content.value)} />
                        ) : (
                            content.value
                        )}
                    </section>
                ) : content.component === 'QueryInput' ? (
                    <QueryInputInViewContent
                        {...props}
                        key={i}
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
    return typeof v === 'object' && v !== null && hasProperty('value')(v) && typeof v.value === 'string'
}
