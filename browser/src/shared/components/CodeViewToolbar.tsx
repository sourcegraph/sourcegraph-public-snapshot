import classNames from 'classnames'
import H from 'history'
import * as React from 'react'
import { ActionNavItemsClassProps, ActionsNavItems } from '../../../../shared/src/actions/ActionsNavItems'
import { ContributionScope } from '../../../../shared/src/api/client/context/context'
import { ContributableMenu } from '../../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../../shared/src/extensions/controller'
import { PlatformContextProps } from '../../../../shared/src/platform/context'
import { TelemetryProps } from '../../../../shared/src/telemetry/telemetryService'
import { FileInfoWithContents } from '../code-hosts/shared/codeViews'
import { OpenDiffOnSourcegraph } from './OpenDiffOnSourcegraph'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'
import { SignInButton } from '../code-hosts/shared/SignInButton'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { isHTTPAuthError } from '../../../../shared/src/backend/fetch'

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
    extends PlatformContextProps<'forceUpdateTooltip' | 'settings' | 'requestGraphQL'>,
        ExtensionsControllerProps,
        TelemetryProps,
        CodeViewToolbarClassProps {
    sourcegraphURL: string

    /**
     * Information about the file or diff the toolbar is displayed on.
     */
    fileInfoOrError: FileInfoWithContents | ErrorLike

    buttonProps?: ButtonProps
    onSignInClose: () => void
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
        {isErrorLike(props.fileInfoOrError) ? (
            isHTTPAuthError(props.fileInfoOrError) ? (
                <SignInButton
                    sourcegraphURL={props.sourcegraphURL}
                    onSignInClose={props.onSignInClose}
                    className={props.actionItemClass}
                    iconClassName={props.actionItemIconClass}
                />
            ) : null
        ) : (
            <>
                {props.fileInfoOrError.baseCommitID && props.fileInfoOrError.baseHasFileContents && (
                    <li className={classNames('code-view-toolbar__item', props.listItemClass)}>
                        <OpenDiffOnSourcegraph
                            ariaLabel="View file diff on Sourcegraph"
                            platformContext={props.platformContext}
                            className={props.actionItemClass}
                            iconClassName={props.actionItemIconClass}
                            openProps={{
                                sourcegraphURL: props.sourcegraphURL,
                                repoName: props.fileInfoOrError.baseRepoName || props.fileInfoOrError.repoName,
                                filePath: props.fileInfoOrError.baseFilePath || props.fileInfoOrError.filePath,
                                rev: props.fileInfoOrError.baseRev || props.fileInfoOrError.baseCommitID,
                                query: {
                                    diff: {
                                        rev: props.fileInfoOrError.baseCommitID,
                                    },
                                },
                                commit: {
                                    baseRev: props.fileInfoOrError.baseRev || props.fileInfoOrError.baseCommitID,
                                    headRev: props.fileInfoOrError.rev || props.fileInfoOrError.commitID,
                                },
                            }}
                        />
                    </li>
                )}{' '}
                {
                    // Only show the "View file" button if we were able to fetch the file contents
                    // from the Sourcegraph instance
                    !props.fileInfoOrError.baseCommitID &&
                        (props.fileInfoOrError.content !== undefined ||
                            props.fileInfoOrError.baseContent !== undefined) && (
                            <li className={classNames('code-view-toolbar__item', props.listItemClass)}>
                                <OpenOnSourcegraph
                                    ariaLabel="View file on Sourcegraph"
                                    className={props.actionItemClass}
                                    iconClassName={props.actionItemIconClass}
                                    openProps={{
                                        sourcegraphURL: props.sourcegraphURL,
                                        repoName: props.fileInfoOrError.repoName,
                                        filePath: props.fileInfoOrError.filePath,
                                        rev: props.fileInfoOrError.rev || props.fileInfoOrError.commitID,
                                    }}
                                />
                            </li>
                        )
                }
            </>
        )}
    </ul>
)
