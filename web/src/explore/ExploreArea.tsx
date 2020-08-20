import * as H from 'history'
import React, { useEffect } from 'react'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import * as GQL from '../../../shared/src/graphql/schema'
import { SettingsCascadeOrError } from '../../../shared/src/settings/settings'
import { eventLogger } from '../tracking/eventLogger'
import { ComponentDescriptor } from '../util/contributions'
import { PatternTypeProps } from '../search'
import { ThemeProps } from '../../../shared/src/theme'
import { PageHeader } from '../components/PageHeader'
import { Breadcrumbs, BreadcrumbsProps } from '../components/Breadcrumbs'
import CompassOutlineIcon from 'mdi-react/CompassOutlineIcon'

/**
 * Properties passed to all section components in the explore area.
 */
export interface ExploreAreaSectionContext
    extends ExtensionsControllerProps,
        ThemeProps,
        Omit<PatternTypeProps, 'setPatternType'> {
    /** The currently authenticated user. */
    authenticatedUser: GQL.IUser | null

    /** The subject whose extensions and settings to display. */
    viewerSubject: Pick<GQL.ISettingsSubject, 'id' | 'viewerCanAdminister'>

    /** The viewer's settings. */
    settingsCascade: SettingsCascadeOrError

    location: H.Location
    history: H.History
}

/** A section shown in the explore area. */
export interface ExploreSectionDescriptor extends ComponentDescriptor<ExploreAreaSectionContext> {}

interface ExploreAreaProps extends ExploreAreaSectionContext, BreadcrumbsProps {
    exploreSections: readonly ExploreSectionDescriptor[]
}

/**
 * The explore area, which shows cards containing summaries and actions from product features. The purpose of it is
 * to expose information at a glance and make it easy to navigate to features (without requiring them to add a link
 * on the space-constrained global nav).
 */
export const ExploreArea: React.FunctionComponent<ExploreAreaProps> = props => {
    useEffect(() => eventLogger.logViewEvent('Explore'), [])

    const context: ExploreAreaSectionContext = {
        extensionsController: props.extensionsController,
        authenticatedUser: props.authenticatedUser,
        viewerSubject: props.viewerSubject,
        settingsCascade: props.settingsCascade,
        isLightTheme: props.isLightTheme,
        location: props.location,
        history: props.history,
        patternType: props.patternType,
    }

    return (
        <div className="explore-area w-100 web-content">
            <Breadcrumbs breadcrumbs={props.breadcrumbs} />
            <div className="container">
                <PageHeader title="Explore" icon={CompassOutlineIcon} />
                {props.exploreSections.map(
                    ({ condition = () => true, render }, index) =>
                        condition(context) && (
                            <div className="mb-5" key={index}>
                                {render(context)}
                            </div>
                        )
                )}
            </div>
        </div>
    )
}
