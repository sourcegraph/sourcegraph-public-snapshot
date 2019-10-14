import classNames from 'classnames'
import H from 'history'
import * as React from 'react'
import { ActionNavItemsClassProps, ActionsNavItems } from '../../../../shared/src/actions/ActionsNavItems'
import { ContributionScope } from '../../../../shared/src/api/client/context/context'
import { ContributableMenu } from '../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { FileInfoWithContents } from '../../libs/code_intelligence/code_views'
import { OpenDiffOnSourcegraph } from './OpenDiffOnSourcegraph'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

export interface ButtonProps {
    className?: string
}

export interface CodeViewToolbarClassProps extends ActionNavItemsClassProps {
    /**
     * Class name for the `<ul>` element wrapping all toolbar items
     */
    className?: string

    /**
     * The scope of this toolbar (e.g., the view component that it is associated with).
     */
    scope?: ContributionScope
}

export interface CodeViewToolbarProps
    extends PlatformContextProps<'forceUpdateTooltip' | 'requestGraphQL'>,
        ExtensionsControllerProps,
        FileInfoWithContents,
        TelemetryProps,
        CodeViewToolbarClassProps {
    sourcegraphURL: string
    onEnabledChange?: (enabled: boolean) => void
    buttonProps?: ButtonProps
    location: H.Location
}

export const CodeViewToolbar: React.FunctionComponent<CodeViewToolbarProps> = props => (
    <ul className={classNames('code-view-toolbar', props.className)}>
        <ActionsNavItems
            {...props}
            listItemClass={classNames('code-view-toolbar__item', props.listItemClass)}
            menu={ContributableMenu.EditorTitle}
            extensionsController={props.extensionsController}
            platformContext={props.platformContext}
            location={props.location}
            scope={props.scope}
        />{' '}
        {props.baseCommitID && props.baseHasFileContents && (
            <li className={classNames('code-view-toolbar__item', props.listItemClass)}>
                <OpenDiffOnSourcegraph
                    ariaLabel="View file diff on Sourcegraph"
                    platformContext={props.platformContext}
                    className={props.actionItemClass}
                    iconClassName={props.actionItemIconClass}
                    openProps={{
                        sourcegraphURL: props.sourcegraphURL,
                        repoName: props.baseRepoName || props.repoName,
                        filePath: props.baseFilePath || props.filePath,
                        rev: props.baseRev || props.baseCommitID,
                        query: {
                            diff: {
                                rev: props.baseCommitID,
                            },
                        },
                        commit: {
                            baseRev: props.baseRev || props.baseCommitID,
                            headRev: props.rev || props.commitID,
                        },
                    }}
                />
            </li>
        )}{' '}
        {// Only show the "View file" button if we were able to fetch the file contents
        // from the Sourcegraph instance
        !props.baseCommitID && (props.content !== undefined || props.baseContent !== undefined) && (
            <li className={classNames('code-view-toolbar__item', props.listItemClass)}>
                <OpenOnSourcegraph
                    ariaLabel="View file on Sourcegraph"
                    className={props.actionItemClass}
                    iconClassName={props.actionItemIconClass}
                    openProps={{
                        sourcegraphURL: props.sourcegraphURL,
                        repoName: props.repoName,
                        filePath: props.filePath,
                        rev: props.rev || props.commitID,
                    }}
                />
            </li>
        )}
    </ul>
)
