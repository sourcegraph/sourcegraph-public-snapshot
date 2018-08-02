import * as H from 'history'
import ArrowLeftDropCircleIcon from 'mdi-react/ArrowLeftDropCircleIcon'
import * as React from 'react'
import { Link } from 'react-router-dom'
import { fromEvent, Subscription } from 'rxjs'
import { filter } from 'rxjs/operators'
import { FileHistoryItem } from './HistoryPopoverContainer'
import { splitPath } from './RepoFileLink'

interface Props {
    fileHistoryList: FileHistoryItem[]
    history: H.History
    location: H.Location
}

export class HistoryPopover extends React.Component<Props> {
    private subscriptions = new Subscription()

    private goToFileViaKeyboard(event: KeyboardEvent): void {
        const currentLocation = this.props.history.location.pathname
        const fileHistoryEntry = this.props.fileHistoryList
        if (!event.shiftKey && !event.altKey && !event.ctrlKey && !event.metaKey) {
            switch (event.key) {
                case '0':
                    if (this.props.fileHistoryList[0] && currentLocation !== fileHistoryEntry[0].url) {
                        this.props.history.push(fileHistoryEntry[0].url)
                    }
                    break
                case '1':
                    if (fileHistoryEntry[1] && currentLocation !== fileHistoryEntry[1].url) {
                        this.props.history.push(`${fileHistoryEntry[1].url}?file-history`)
                    }
                    break
                case '2':
                    if (fileHistoryEntry[2] && currentLocation !== fileHistoryEntry[2].url) {
                        this.props.history.push(`${fileHistoryEntry[2].url}?file-history`)
                    }
                    break
                case '3':
                    if (fileHistoryEntry[3] && currentLocation !== fileHistoryEntry[3].url) {
                        this.props.history.push(`${fileHistoryEntry[3].url}?file-history`)
                    }
                    break
                case '4':
                    if (fileHistoryEntry[4] && currentLocation !== fileHistoryEntry[4].url) {
                        this.props.history.push(`${fileHistoryEntry[4].url}?file-history`)
                    }
                    break
                case '5':
                    if (fileHistoryEntry[5] && currentLocation !== fileHistoryEntry[5].url) {
                        this.props.history.push(`${fileHistoryEntry[5].url}?file-history`)
                    }
                    break
                case '6':
                    if (fileHistoryEntry[6] && currentLocation !== fileHistoryEntry[6].url) {
                        this.props.history.push(`${fileHistoryEntry[6].url}?file-history`)
                    }
                    break
                case '7':
                    if (fileHistoryEntry[7] && currentLocation !== fileHistoryEntry[7].url) {
                        this.props.history.push(`${fileHistoryEntry[7].url}?file-history`)
                    }
                    break
                case '8':
                    if (fileHistoryEntry[8] && currentLocation !== fileHistoryEntry[8].url) {
                        this.props.history.push(`${fileHistoryEntry[8].url}?file-history`)
                    }
                    break
                case '9':
                    if (fileHistoryEntry[9] && currentLocation !== fileHistoryEntry[9].url) {
                        this.props.history.push(`${fileHistoryEntry[9].url}?file-history`)
                    }
                    break
            }
        }
    }

    public componentDidMount(): void {
        const numberKeyUp = fromEvent<KeyboardEvent>(window, 'keyup').pipe(
            filter(event => event.keyCode >= 48 && event.keyCode <= 57)
        )
        this.subscriptions.add(numberKeyUp.subscribe(event => this.goToFileViaKeyboard(event)))
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element {
        let prevRepo: string
        return (
            <div className="history-popover">
                {this.props.fileHistoryList.length > 0 ? (
                    this.props.fileHistoryList.map((file, i) => {
                        let label: JSX.Element | number = (
                            <code className="history-popover__label history-popover__label-number">{i}</code>
                        )
                        if (file.url === location.pathname) {
                            label = <ArrowLeftDropCircleIcon className="history-popover__label icon-inline" />
                        }
                        const [fileBase, fileName] = splitPath(file.filePath)

                        const element =
                            prevRepo === file.repoPath ? (
                                // If one of the 10 most recent files viewed, don't update history since it's very jarring in the UI.
                                // If it's an older entry, it probably makes sense to add to history since it's probably a new context.
                                <Link to={i <= 10 ? `${file.url}?file-history` : file.url} key={file.filePath + i}>
                                    <li className=" history-popover__list-group-item list-group-item">
                                        {fileName}{' '}
                                        {fileBase.length > 0 ? (
                                            <span className="history-popover__description text-muted">{`- ${fileBase}`}</span>
                                        ) : (
                                            ''
                                        )}{' '}
                                        {i < 10 && label}
                                    </li>
                                </Link>
                            ) : (
                                <div key={file.filePath + i}>
                                    <div className="history-popover__header history-popover__list-group-item list-group-item">
                                        {file.repoPath}
                                    </div>
                                    <Link to={i <= 10 ? `${file.url}?file-history` : file.url}>
                                        <li className="history-popover__list-group-item list-group-item">
                                            {fileName}{' '}
                                            {fileBase.length > 0 ? (
                                                <span className="history-popover__description text-muted">{`- ${fileBase}`}</span>
                                            ) : (
                                                ''
                                            )}{' '}
                                            {i < 10 && label}
                                        </li>
                                    </Link>
                                </div>
                            )
                        prevRepo = file.repoPath
                        return element
                    })
                ) : (
                    <div className="text-muted small">No file history yet</div>
                )}
            </div>
        )
    }
}
