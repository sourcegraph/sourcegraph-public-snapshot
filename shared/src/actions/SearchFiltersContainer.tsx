import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import { switchMap } from 'rxjs/operators'
import { ContributionScope } from '../api/client/context/context'
import { Contributions } from '../api/protocol'
import { SearchFilters } from '../api/protocol'
import { ExtensionsControllerProps } from '../extensions/controller'

export interface SearchFiltersProps extends ExtensionsControllerProps {
    scope?: ContributionScope
}
interface SearchFiltersContainerProps extends SearchFiltersProps {
    /**
     * Called with the array of contributed items to produce the rendered component. If not set, it doesn't render.
     */
    render: (items: SearchFilters[]) => React.ReactElement<any>

    /**
     * If set, it is rendered when there are no contributed items for this menu. Use null to render nothing when
     * empty.
     */
    empty?: React.ReactElement<any> | null
}

export interface SearchFiltersContainerState {
    /** The contributions, merged from all extensions, or undefined before the initial emission. */
    contributions?: Contributions
}

/** Displays the search filters in a container, with a wrapper and/or empty element. */
export class SearchFiltersContainer extends React.PureComponent<
    SearchFiltersContainerProps,
    SearchFiltersContainerState
> {
    public state: SearchFiltersContainerState = {}

    private scopeChanges = new Subject<ContributionScope | undefined>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.subscriptions.add(
            this.scopeChanges
                .pipe(switchMap(scope => this.props.extensionsController.services.contribution.getContributions(scope)))
                .subscribe(contributions => this.setState({ contributions }))
        )
        this.scopeChanges.next(this.props.scope)
    }

    public componentDidUpdate(prevProps: SearchFiltersContainerProps): void {
        if (prevProps.scope !== this.props.scope) {
            this.scopeChanges.next(this.props.scope)
        }
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        if (!this.state.contributions) {
            return null // loading
        }
        return this.props.render(this.state.contributions.searchFilters || [])
    }
}
