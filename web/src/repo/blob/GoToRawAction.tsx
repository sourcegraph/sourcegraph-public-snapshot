import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import * as React from 'react'
import { encodeRepoRev } from '../../../../shared/src/util/url'

interface Props {
    repoName: string
    rev?: string
    filePath: string
}

/**
 * A repository header action that replaces the blob in the URL with the raw URL.
 */
export class GoToRawAction extends React.PureComponent<Props> {
    public render(): JSX.Element {
        const to = `/${encodeRepoRev(this.props)}/-/raw/${this.props.filePath}`
        return (
            <a href={to} className="nav-link" data-tooltip="Raw (download file)" download={true}>
                <FileDownloadIcon className="icon-inline" />
            </a>
        )
    }
}
