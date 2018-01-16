import * as React from 'react'
import { interval } from 'rxjs/observable/interval'
import { catchError } from 'rxjs/operators/catchError'
import { delay } from 'rxjs/operators/delay'
import { filter } from 'rxjs/operators/filter'
import { switchMap } from 'rxjs/operators/switchMap'
import { take } from 'rxjs/operators/take'
import { Subscription } from 'rxjs/Subscription'
import { SiteFlags } from '../site'
import { refreshSiteFlags, siteFlags } from '../site/backend'
import { NeedsRepositoryConfigurationAlert } from '../site/NeedsRepositoryConfigurationAlert'
import { RepositoriesCloningAlert } from '../site/RepositoriesCloningAlert'

interface Props {
    isSiteAdmin: boolean
}

interface State {
    siteFlags?: SiteFlags
}

/**
 * Fetches and displays relevant global alerts at the top of the page
 */
export class GlobalAlerts extends React.PureComponent<Props, State> {
    public state: State = {}

    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(siteFlags.subscribe(siteFlags => this.setState({ siteFlags })))

        if (this.props.isSiteAdmin) {
            // Refresh site flags periodically while repositories are cloning.
            this.subscriptions.add(
                siteFlags
                    .pipe(
                        filter(
                            ({ repositoriesCloning }) =>
                                repositoriesCloning.totalCount !== null && repositoriesCloning.totalCount > 0
                        )
                    )
                    .pipe(delay(5000), switchMap(refreshSiteFlags))
                    .subscribe()
            )

            // Also periodically fetch (but less often) always.
            this.subscriptions.add(
                interval(5000)
                    .pipe(take(3), delay(3000), switchMap(refreshSiteFlags), catchError(() => []))
                    .subscribe()
            )
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (
            this.state.siteFlags &&
            this.state.siteFlags.repositoriesCloning &&
            this.state.siteFlags.repositoriesCloning.totalCount !== null &&
            this.state.siteFlags.repositoriesCloning.totalCount > 0
        ) {
            return <RepositoriesCloningAlert repositoriesCloning={this.state.siteFlags.repositoriesCloning} />
        }
        if (this.state.siteFlags && this.state.siteFlags.needsRepositoryConfiguration) {
            return <NeedsRepositoryConfigurationAlert />
        }
        return null
    }
}
