import React, { useMemo } from 'react'
import { Section } from '../../../shared/src/components/sections/Sections'
import { Notices } from '../global/Notices'
import { SettingsCascadeProps, SettingsCascadeOrError } from '../../../shared/src/settings/settings'
import { isErrorLike } from '../../../shared/src/util/errors'
import { Settings } from '../schema/settings.schema'
import { HomeRepositories } from './HomeRepositories'

/**
 * Reports the value of the feature flag for using the new HomeSections component on the homepage.
 */
export const isHomeSectionsEnabled = (settingsCascade: SettingsCascadeOrError<Settings>): boolean =>
    settingsCascade.final !== null &&
    !isErrorLike(settingsCascade.final) &&
    Boolean(settingsCascade.final.experimentalFeatures?.homeSections)

interface Props extends SettingsCascadeProps<Settings> {}

type SectionID = 'Welcome' | 'Notices' | 'Repositories'

/**
 * A list of collapsible sections on the homepage.
 */
export const HomeSections: React.FunctionComponent<Props> = ({ settingsCascade }) => {
    const sections = useMemo<Section<SectionID>[]>(
        () => [
            { id: 'Welcome', label: 'Welcome' },
            { id: 'Notices', label: 'Notices' },
            { id: 'Repositories', label: 'Repositories' },
        ],
        []
    )
    return (
        <div className="home-sections">
            <Notices className="my-3" location="home" settingsCascade={settingsCascade} />
            <HomeRepositories
                settings={
                    settingsCascade.final !== null && !isErrorLike(settingsCascade.final) ? settingsCascade.final : {}
                }
            />
        </div>
    )
}
