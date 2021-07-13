import classnames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo } from 'react'
import { Link } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'

import { AuthenticatedUser } from '../../../../auth'
import { HeroPage } from '../../../../components/HeroPage'
import { Page } from '../../../../components/Page'
import { PageTitle } from '../../../../components/PageTitle'
import { isLangStatsInsight, isSearchBasedInsight } from '../../../core/types'
import { useInsightSubjects } from '../../../hooks/use-insight-subjects/use-insight-subjects'
import { useInsight } from '../../../hooks/use-insight/use-insight'

import { EditLangStatsInsight } from './components/EditLangStatsInsight'
import { EditSearchBasedInsight } from './components/EditSearchInsight'
import styles from './EditInsightPage.module.scss'
import { useHandleSubmit } from './hooks/use-handle-submit'

export interface EditInsightPageProps extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {
    /** Normalized insight id <type insight>.insight.<name of insight> */
    insightID: string

    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     * */
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'organizations' | 'username'>
}

export const EditInsightPage: React.FunctionComponent<EditInsightPageProps> = props => {
    const { insightID, settingsCascade, authenticatedUser, platformContext } = props

    const subjects = useInsightSubjects({ settingsCascade })
    const insight = useInsight({ settingsCascade, insightId: insightID })
    const { handleEditInsightSubmit } = useHandleSubmit({
        originalInsight: insight,
        settingsCascade,
        platformContext,
    })

    const finalSettings = useMemo(() => {
        if (!insight) {
            return settingsCascade.final ?? {}
        }

        const newSettings: Settings = { ...settingsCascade.final }

        // Final settings used below as a store of all existing insights
        // Usually we have validation for title of insight because user can't
        // have two insights with the same name/id.
        // In edit mode we should allow users to have insight with id (camelCase(insight title))
        // which already exists in setting store. For turning off this id/title validation
        // we are removing current insight from final settings.
        delete newSettings[insightID]

        return newSettings
    }, [settingsCascade.final, insight, insightID])

    if (!insight) {
        return (
            <HeroPage
                icon={MapSearchIcon}
                title="Oops, we couldn't find insight"
                subtitle={
                    <span>
                        We couldn't find that insight. Try to find the insight with ID:{' '}
                        <code className="badge badge-secondary">{insightID}</code> in your{' '}
                        <Link to={`/users/${authenticatedUser?.username}/settings`}>user or org settings</Link>
                    </span>
                }
            />
        )
    }

    return (
        <Page className={classnames('col-10', styles.creationPage)}>
            <PageTitle title="Edit code insight" />

            <div className="mb-5">
                <h2>Edit insight</h2>

                <p className="text-muted">
                    Insights analyze your code based on any search query.{' '}
                    <a href="https://docs.sourcegraph.com/code_insights" target="_blank" rel="noopener">
                        Learn more.
                    </a>
                </p>
            </div>

            {isSearchBasedInsight(insight) && (
                <EditSearchBasedInsight
                    insight={insight}
                    finalSettings={finalSettings}
                    subjects={subjects}
                    onSubmit={handleEditInsightSubmit}
                />
            )}

            {isLangStatsInsight(insight) && (
                <EditLangStatsInsight
                    insight={insight}
                    finalSettings={finalSettings}
                    subjects={subjects}
                    onSubmit={handleEditInsightSubmit}
                />
            )}
        </Page>
    )
}
