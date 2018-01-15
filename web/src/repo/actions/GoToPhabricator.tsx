import PhabricatorIcon from '@sourcegraph/icons/lib/Phabricator'
import * as H from 'history'
import * as React from 'react'
import { catchError } from 'rxjs/operators/catchError'
import { switchMap } from 'rxjs/operators/switchMap'
import { tap } from 'rxjs/operators/tap'
import { Subject } from 'rxjs/Subject'
import { Subscription } from 'rxjs/Subscription'
import { parseBrowserRepoURL } from '..'
import { eventLogger } from '../../tracking/eventLogger'
import { fetchPhabricatorRepo } from '../backend'

interface PhabricatorRepo {
    callsign: string
    url: string
}

interface Props {
    repo: string
    location: H.Location
}

interface State {
    phabRepo?: PhabricatorRepo | null
}

/**
 * A repository header action that goes to the corresponding URL on Phabricator.
 */
export class GoToPhabricatorAction extends React.PureComponent<Props, State> {
    public state: State = {}

    private repoChanges = new Subject<string>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.repoChanges
                .pipe(
                    tap(() => {
                        if (this.state.phabRepo) {
                            this.setState({ phabRepo: null })
                        }
                    }),
                    switchMap(repo =>
                        fetchPhabricatorRepo({ repoPath: repo }).pipe(
                            catchError(err => {
                                console.error(err)
                                return []
                            })
                        )
                    )
                )
                .subscribe(phabRepo => this.setState({ phabRepo }), err => console.error(err))
        )

        this.repoChanges.next(this.props.repo)
    }

    public componentWillReceiveProps(props: Props): void {
        if (props.repo !== this.props.repo) {
            this.repoChanges.next(props.repo)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.phabRepo) {
            return null
        }

        const { filePath } = parseBrowserRepoURL(this.props.location.pathname)

        return (
            <a
                className="btn btn-link btn-sm composite-container__header-action"
                onClick={onClick}
                href={urlToPhabricator(this.state.phabRepo, filePath)}
                data-tooltip="View on Phabricator"
            >
                <PhabricatorIcon className="icon-inline" />
                {/* Most people won't recognize the Phabricator icon, so include the text label here, too. */}
                <span className="composite-container__header-action-text">Phabricator</span>
            </a>
        )
    }
}

function onClick(): void {
    eventLogger.log('OpenInCodeHostClicked')
}

// TODO(sqs): include rev
function urlToPhabricator(phabRepo: PhabricatorRepo, filePath: string | undefined): string {
    return `${phabRepo.url}/source/${phabRepo.callsign}/browse/${filePath || ''}`
}
