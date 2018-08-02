import H from 'history'
import { isEqual } from 'lodash'
import * as React from 'react'
import { fromEvent, Subject, Subscription } from 'rxjs'
import { distinctUntilChanged, filter } from 'rxjs/operators'
import { parseBrowserRepoURL } from '../repo'
import { HistoryPopover } from './HistoryPopover'
import { PopoverButton } from './PopoverButton'
import { displayRepoPath } from './RepoFileLink'

interface Props {
    location: H.Location
    history: H.History
}

export interface FileHistoryItem {
    repoPath: string
    url: string
    filePath: string
}

interface State {
    fileHistoryList: FileHistoryItem[]
}

const FILE_HISTORY_KEY = 'fileHistoryList'

export class HistoryPopoverContainer extends React.Component<Props, State> {
    public subscriptions = new Subscription()
    public locationUpdates = new Subject<Props>()
    public state = { fileHistoryList: JSON.parse(localStorage.getItem(FILE_HISTORY_KEY) || '{}') }

    public updateFileHistory(location: H.Location): void {
        const loc = parseBrowserRepoURL(location.pathname + location.search + location.hash)
        const filePath = loc.filePath
        const repoPath = loc.repoPath
        const url = location.pathname
        const currentHistory = localStorage.getItem(FILE_HISTORY_KEY)
        let newHistoryJson: FileHistoryItem[]
        const queryParams = new URLSearchParams(this.props.location.search)

        if (filePath && url && !queryParams.has('file-history')) {
            const newEntry: FileHistoryItem = {
                repoPath: displayRepoPath(repoPath),
                filePath,
                url,
            }
            if (!currentHistory) {
                localStorage.setItem(FILE_HISTORY_KEY, JSON.stringify([newEntry]))
                this.setState({ fileHistoryList: [newEntry] })
            } else {
                newHistoryJson = JSON.parse(currentHistory)
                if (newEntry.url !== newHistoryJson[0].url) {
                    newHistoryJson = [newEntry, ...newHistoryJson].slice(0, 100)
                    localStorage.setItem(FILE_HISTORY_KEY, JSON.stringify(newHistoryJson))
                    this.setState({ fileHistoryList: newHistoryJson })
                }
            }
        }

        if (queryParams.has('file-history')) {
            queryParams.delete('file-history')
            this.props.history.replace({ search: queryParams.toString() })
        }
    }

    public componentDidMount(): void {
        this.subscriptions.add(
            this.locationUpdates
                .pipe(distinctUntilChanged((a, b) => isEqual(a.location, b.location)))
                .subscribe(props => {
                    this.updateFileHistory(props.location)
                })
        )

        this.subscriptions.add(
            fromEvent<StorageEvent>(window, 'storage')
                .pipe(filter(e => e.key === FILE_HISTORY_KEY))
                .subscribe(e => {
                    if (e.newValue) {
                        this.setState({ fileHistoryList: JSON.parse(e.newValue) })
                    }
                })
        )

        this.locationUpdates.next(this.props)
    }

    public componentDidUpdate(prevProps: Props, prevState: State): void {
        if (
            prevProps.location.pathname !== this.props.location.pathname ||
            prevState.fileHistoryList !== this.state.fileHistoryList
        ) {
            this.locationUpdates.next(this.props)
        }
    }
    public render(): JSX.Element {
        return (
            <>
                <PopoverButton
                    globalKeyBinding="h"
                    popoverElement={<HistoryPopover fileHistoryList={this.state.fileHistoryList} {...this.props} />}
                    placement="bottom-end"
                    className="history-popover__link-color"
                >
                    History
                </PopoverButton>
            </>
        )
    }
}
