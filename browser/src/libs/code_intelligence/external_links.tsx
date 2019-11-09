import classNames from 'classnames'
import { isEqual } from 'lodash'
import * as React from 'react'
import { render } from 'react-dom'
import { Observable, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, map, switchMap, first } from 'rxjs/operators'
import { SourcegraphIconButton } from '../../shared/components/Button'
import { DEFAULT_SOURCEGRAPH_URL } from '../../shared/util/context'
import { CodeHost, CodeHostContext } from './code_intelligence'
import { useObservable } from '../../../../shared/src/util/useObservable'

export interface ViewOnSourcegraphButtonClassProps {
    className?: string
    iconClassName?: string
}

interface ViewOnSourcegraphButtonProps extends ViewOnSourcegraphButtonClassProps {
    context: CodeHostContext
    sourcegraphURL: Observable<URL>
    ensureRepoExists: (context: CodeHostContext) => Observable<boolean>
    onConfigureSourcegraphClick?: () => void
}

const ViewOnSourcegraphButton: React.FunctionComponent<ViewOnSourcegraphButtonProps> = ({
    ensureRepoExists,
    context,
    sourcegraphURL,
    className,
    iconClassName,
    onConfigureSourcegraphClick,
}) => {
    const repoExists = useObservable(React.useMemo(() => ensureRepoExists(context), [context, ensureRepoExists]))
    const sourcegraphBaseURL = useObservable(sourcegraphURL)

    if (!repoExists) {
        return null
    }

    if (!sourcegraphBaseURL) {
        return null
    }

    // If repo doesn't exist and the instance is sourcegraph.com, prompt
    // user to configure Sourcegraph.
    if (!repoExists && sourcegraphBaseURL.href === DEFAULT_SOURCEGRAPH_URL.href && onConfigureSourcegraphClick) {
        return (
            <SourcegraphIconButton
                label="Configure Sourcegraph"
                ariaLabel="Install Sourcegraph for search and code intelligence on private instance"
                className={classNames('open-on-sourcegraph', className)}
                iconClassName={classNames('open-on-sourcegraph__icon--muted', iconClassName)}
                onClick={onConfigureSourcegraphClick}
            />
        )
    }

    const rev = context.rev ? `@${context.rev}` : ''
    const iconURL = new URL(`${context.rawRepoName}${rev}`, sourcegraphBaseURL).href
    return (
        <SourcegraphIconButton
            url={iconURL}
            ariaLabel="View repository on Sourcegraph"
            className={classNames('open-on-sourcegraph', className)}
            iconClassName={iconClassName}
        />
    )
}

export const renderViewContextOnSourcegraph = ({
    sourcegraphURL,
    getContext,
    ensureRepoExists,
    viewOnSourcegraphButtonClassProps,
    onConfigureSourcegraphClick,
}: {
    sourcegraphURL: Observable<URL>
    ensureRepoExists: ViewOnSourcegraphButtonProps['ensureRepoExists']
    onConfigureSourcegraphClick?: ViewOnSourcegraphButtonProps['onConfigureSourcegraphClick']
} & Required<Pick<CodeHost, 'getContext'>> &
    Pick<CodeHost, 'viewOnSourcegraphButtonClassProps'>) => (mount: HTMLElement): void => {
    render(
        <ViewOnSourcegraphButton
            {...viewOnSourcegraphButtonClassProps}
            context={getContext()}
            sourcegraphURL={sourcegraphURL}
            ensureRepoExists={ensureRepoExists}
            onConfigureSourcegraphClick={onConfigureSourcegraphClick}
        />,
        mount
    )
}
