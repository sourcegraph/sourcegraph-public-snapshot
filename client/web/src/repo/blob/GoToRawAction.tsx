import * as React from 'react'

import FileDownloadOutlineIcon from 'mdi-react/FileDownloadOutlineIcon'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { encodeRepoRevision, RepoSpec, RevisionSpec, FileSpec } from '@sourcegraph/shared/src/util/url'
import { Icon } from '@sourcegraph/wildcard'

import { RepoHeaderActionAnchor } from '../components/RepoHeaderActions'
import { RepoHeaderContext } from '../RepoHeader'

interface Props extends RepoSpec, Partial<RevisionSpec>, FileSpec, RepoHeaderContext, TelemetryProps {}

/**
 * A repository header action that replaces the blob in the URL with the raw URL.
 */
export class GoToRawAction extends React.PureComponent<Props> {
    private onClick(): void {
        this.props.telemetryService.log('RawFileDownload', {
            repoName: this.props.repoName,
            filePath: this.props.filePath,
        })
    }

    public render(): JSX.Element {
        const to = `/${encodeRepoRevision(this.props)}/-/raw/${this.props.filePath}`
        const descriptiveText = 'Raw (download file)'

        if (this.props.actionType === 'dropdown') {
            return (
                <RepoHeaderActionAnchor
                    to={to}
                    target="_blank"
                    file={true}
                    onClick={this.onClick.bind(this)}
                    download={true}
                >
                    <Icon as={FileDownloadOutlineIcon} />
                    <span>{descriptiveText}</span>
                </RepoHeaderActionAnchor>
            )
        }

        return (
            <RepoHeaderActionAnchor
                to={to}
                target="_blank"
                onClick={this.onClick.bind(this)}
                className="btn-icon"
                data-tooltip={descriptiveText}
                aria-label={descriptiveText}
                download={true}
            >
                <Icon as={FileDownloadOutlineIcon} />
            </RepoHeaderActionAnchor>
        )
    }
}
