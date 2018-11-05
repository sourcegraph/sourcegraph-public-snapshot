import {
    ConfigurationCascadeOrError,
    ConfigurationSubject,
    Settings,
} from '@sourcegraph/extensions-client-common/lib/settings'
import H from 'history'
import * as React from 'react'
import { Subject, Subscription } from 'rxjs'
import * as GQL from '../backend/graphqlschema'
import { ExtensionsControllerProps } from '../extensions/ExtensionsClientCommonContext'
import { ComponentDescriptor } from '../util/contributions'

/**
 * Properties passed to all section components in the explore area.
 */
export interface ExploreAreaSectionContext extends ExtensionsControllerProps {
    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null

    /** The subject whose extensions and configuration to display. */
    viewerSubject: Pick<GQL.IConfigurationSubject, 'id' | 'viewerCanAdminister'>

    /** The viewer's configuration. */
    configurationCascade: ConfigurationCascadeOrError<ConfigurationSubject, Settings>

    isLightTheme: boolean
    location: H.Location
    history: H.History
}

/** A section shown in the explore area. */
export interface ExploreSectionDescriptor extends ComponentDescriptor<ExploreAreaSectionContext> {}

interface ExploreAreaProps extends ExploreAreaSectionContext {
    exploreSections: ReadonlyArray<ExploreSectionDescriptor>
}

const LOADING: 'loading' = 'loading'

interface ExploreAreaState {}

/**
 * The explore area, which shows cards containing summaries and actions from product features. The purpose of it is
 * to expose information at a glance and make it easy to navigate to features (without requiring them to add a link
 * on the space-constrained global nav).
 */
export class ExploreArea extends React.Component<ExploreAreaProps, ExploreAreaState> {
    public state: ExploreAreaState = {
        subjectOrError: LOADING,
    }

    private componentUpdates = new Subject<ExploreAreaProps>()
    private subscriptions = new Subscription()

    public componentDidMount(): void {
        this.componentUpdates.next(this.props)
    }

    public componentWillReceiveProps(props: ExploreAreaProps): void {
        this.componentUpdates.next(props)
    }

    public componentWillUnmount(): void {
        this.subscriptions.unsubscribe()
    }

    public render(): JSX.Element | null {
        const context: ExploreAreaSectionContext = {
            extensionsController: this.props.extensionsController,
            authenticatedUser: this.props.authenticatedUser,
            viewerSubject: this.props.viewerSubject,
            configurationCascade: this.props.configurationCascade,
            isLightTheme: this.props.isLightTheme,
            location: this.props.location,
            history: this.props.history,
        }

        return (
            <div className="explore-area container my-3">
                <h1>Explore</h1>
                {this.props.exploreSections.map(
                    ({ condition = () => true, render }, i) =>
                        condition(context) && (
                            <div className="mb-5" key={i}>
                                {render(context)}
                            </div>
                        )
                )}
            </div>
        )
    }
}
