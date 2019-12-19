import FileDownloadIcon from 'mdi-react/FileDownloadIcon'
import * as React from 'react'
import { LinkOrButton } from '../../../../shared/src/components/LinkOrButton'
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
        const to = `/${encodeRepoRev(this.props.repoName, this.props.rev)}/-/raw/${this.props.filePath}`
        return (
            <LinkOrButton to={to} data-tooltip="Raw (download file)">
                <FileDownloadIcon className="icon-inline" />
            </LinkOrButton>
        )
    }
}
