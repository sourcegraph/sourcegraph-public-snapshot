import * as React from 'react'

import classNames from 'classnames'
import type * as H from 'history'

import { ContributableMenu } from '@sourcegraph/client-api'
import { type ErrorLike, isErrorLike } from '@sourcegraph/common'
import { isHTTPAuthError } from '@sourcegraph/http-client'
import { type ActionNavItemsClassProps, ActionsNavItems } from '@sourcegraph/shared/src/actions/ActionsNavItems'
import type { ContributionScope } from '@sourcegraph/shared/src/api/extension/api/context/context'
import type { ExtensionsControllerProps } from '@sourcegraph/shared/src/extensions/controller'
import type { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import type { TelemetryV2Props } from '@sourcegraph/shared/src/telemetry'
import type { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'

import type { DiffOrBlobInfo, FileInfoWithContent } from '../code-hosts/shared/codeHost'
import { SignInButton } from '../code-hosts/shared/SignInButton'
import { defaultRevisionToCommitID } from '../code-hosts/shared/util/fileInfo'

import { OpenDiffOnSourcegraph } from './OpenDiffOnSourcegraph'
import { OpenOnSourcegraph } from './OpenOnSourcegraph'

import styles from './CodeViewToolbar.module.scss'

export interface ButtonProps {
    listItemClass?: string
    actionItemClass?: string
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
    extends PlatformContextProps<'settings' | 'requestGraphQL'>,
        ExtensionsControllerProps,
        TelemetryProps,
        TelemetryV2Props,
        CodeViewToolbarClassProps {
    sourcegraphURL: string

    /**
     * Information about the file or diff the toolbar is displayed on.
     */
    fileInfoOrError: DiffOrBlobInfo<FileInfoWithContent> | ErrorLike

    /**
     * Code-view specific className overrides.
     */
    buttonProps?: ButtonProps
    onSignInClose: () => void
    location: H.Location
    hideActions?: boolean
}

export const CodeViewToolbar: React.FunctionComponent<React.PropsWithChildren<CodeViewToolbarProps>> = props => (
    <ul className={classNames(styles.codeViewToolbar, props.className)} data-testid="code-view-toolbar">
        {!props.hideActions && props.extensionsController !== null && (
            <ActionsNavItems
                {...props}
                listItemClass={classNames(styles.item, props.buttonProps?.listItemClass ?? props.listItemClass)}
                actionItemClass={classNames(props.buttonProps?.actionItemClass ?? props.actionItemClass)}
                menu={ContributableMenu.EditorTitle}
                extensionsController={props.extensionsController}
                platformContext={props.platformContext}
                location={props.location}
                scope={props.scope}
            />
        )}{' '}
        {isErrorLike(props.fileInfoOrError) ? (
            isHTTPAuthError(props.fileInfoOrError) ? (
                <SignInButton
                    sourcegraphURL={props.sourcegraphURL}
                    onSignInClose={props.onSignInClose}
                    className={classNames(props.buttonProps?.actionItemClass ?? props.actionItemClass)}
                    iconClassName={props.actionItemIconClass}
                />
            ) : null
        ) : (
            <>
                {!('blob' in props.fileInfoOrError) && props.fileInfoOrError.head && props.fileInfoOrError.base && (
                    <li className={classNames(styles.item, props.buttonProps?.listItemClass ?? props.listItemClass)}>
                        <OpenDiffOnSourcegraph
                            ariaLabel="View file diff on Sourcegraph"
                            platformContext={props.platformContext}
                            className={classNames(props.buttonProps?.actionItemClass ?? props.actionItemClass)}
                            iconClassName={props.actionItemIconClass}
                            openProps={{
                                sourcegraphURL: props.sourcegraphURL,
                                repoName: props.fileInfoOrError.base.repoName,
                                filePath: props.fileInfoOrError.base.filePath,
                                revision: defaultRevisionToCommitID(props.fileInfoOrError.base).revision,
                                commit: {
                                    baseRev: defaultRevisionToCommitID(props.fileInfoOrError.base).revision,
                                    headRev: defaultRevisionToCommitID(props.fileInfoOrError.head).revision,
                                },
                            }}
                        />
                    </li>
                )}{' '}
                {
                    // Only show the "View file" button if we were able to fetch the file contents
                    // from the Sourcegraph instance
                    'blob' in props.fileInfoOrError && props.fileInfoOrError.blob.content !== undefined && (
                        <li
                            className={classNames(
                                styles.item,
                                props.buttonProps?.actionItemClass ?? props.listItemClass
                            )}
                        >
                            <OpenOnSourcegraph
                                ariaLabel="View file on Sourcegraph"
                                className={classNames(props.buttonProps?.actionItemClass ?? props.actionItemClass)}
                                iconClassName={props.actionItemIconClass}
                                openProps={{
                                    sourcegraphURL: props.sourcegraphURL,
                                    repoName: props.fileInfoOrError.blob.repoName,
                                    filePath: props.fileInfoOrError.blob.filePath,
                                    revision: defaultRevisionToCommitID(props.fileInfoOrError.blob).revision,
                                }}
                            />
                        </li>
                    )
                }
            </>
        )}
    </ul>
)
