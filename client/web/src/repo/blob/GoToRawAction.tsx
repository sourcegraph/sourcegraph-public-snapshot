import FileDownloadOutlineIcon from 'mdi-react/FileDownloadOutlineIcon'
import * as React from 'react'

import { encodeRepoRevision, RepoSpec, RevisionSpec, FileSpec } from '@sourcegraph/shared/src/util/url'

import { RepoHeaderContext } from '../RepoHeader'

interface Props extends RepoSpec, Partial<RevisionSpec>, FileSpec, RepoHeaderContext {}

/**
 * A repository header action that replaces the blob in the URL with the raw URL.
 */
export class GoToRawAction extends React.PureComponent<Props> {
    public render(): JSX.Element {
        const to = `/${encodeRepoRevision(this.props)}/-/raw/${this.props.filePath}`

        if (this.props.actionType === 'dropdown') {
            return (
                <a href={to} className="nav-link repo-header__file-action" download={true}>
                    <FileDownloadOutlineIcon className="icon-inline" />
                    <span>Raw (download file)</span>
                </a>
            )
        }

        return (
            <a href={to} className="nav-link" data-tooltip="Raw (download file)" download={true}>
                <FileDownloadOutlineIcon className="icon-inline" />
            </a>
        )
    }
}
