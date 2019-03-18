import { formatDistance } from 'date-fns'
import { isEqual, truncate } from 'lodash'
import * as React from 'react'
import ReactDOM from 'react-dom'
import { Link } from 'react-router-dom'
import { merge, of, Subject, Subscription } from 'rxjs'
import { catchError, delay, switchMap, takeUntil } from 'rxjs/operators'

import { AbsoluteRepoFile } from '../..'
import { asError, ErrorLike, isErrorLike } from '../../../util/errors'
import { fetchBlameFile2 } from './backend'

const LOADING: 'loading' = 'loading'

/** The time in ms after which to show a loader if the result has not returned yet */
const LOADER_DELAY = 100

interface BlameLineProps extends AbsoluteRepoFile {
    line: number
    portalID: string
}

interface BlameLineState {
    content: string
    href: string
}

export class BlameLine extends React.Component<BlameLineProps, BlameLineState> {
    public state: BlameLineState = {
        content: '',
        href: '#',
    }

    private portal: Element | null = null

    /** Emits with the latest Props on every componentDidUpdate and on componentDidMount */
    private componentUpdates = new Subject<BlameLineProps>()

    /** Subscriptions to be disposed on unmount */
    private subscriptions = new Subscription()

    constructor(props: BlameLineProps) {
        super(props)

        this.subscriptions.add(
            this.componentUpdates
                .pipe(
                    switchMap(props => {
                        const blameFetch = fetchBlameFile2({
                            repoPath: props.repoPath,
                            commitID: props.commitID,
                            filePath: props.filePath,
                            line: props.line,
                        }).pipe(catchError((error): ErrorLike[] => [asError(error)]))

                        return merge(
                            blameFetch,
                            of(LOADING).pipe(
                                delay(LOADER_DELAY),
                                takeUntil(blameFetch)
                            )
                        )
                    })
                )
                .subscribe(hunksOrError => {
                    let content: string
                    let href = '#'

                    if (hunksOrError === LOADING) {
                        content = 'loading ◌'
                    } else if (isErrorLike(hunksOrError)) {
                        content = 'blame error: ' + hunksOrError.message
                    } else {
                        const hunk = hunksOrError[0]
                        // rev is actually the commitID
                        const commitID = truncate(hunk.rev, { length: 7, omission: '' })
                        const message = truncate(hunk.message, { length: 80, omission: '…' })
                        const timeSince = formatDistance(hunk.author.date, new Date(), { addSuffix: true })
                        content = `${hunk.author.person.name}, ${timeSince} • ${message} ${commitID}`
                        href = hunk.commit.url
                    }

                    this.setState({ content, href })
                })
        )
    }

    public shouldComponentUpdate(nextProps: Readonly<BlameLineProps>, nextState: Readonly<BlameLineState>): boolean {
        return !isEqual(this.props, nextProps) || !isEqual(this.state, nextState)
    }

    public componentDidMount(): void {
        this.portal = document.getElementById(this.props.portalID)

        this.componentUpdates.next(this.props)
    }

    public componentDidUpdate(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): React.ReactPortal | null {
        if (!this.portal) {
            return null
        }

        return ReactDOM.createPortal(
            <Link className="blame" to={this.state.href}>
                <span className="blame__contents" data-contents={this.state.content} />
            </Link>,
            this.portal
        )
    }
}
