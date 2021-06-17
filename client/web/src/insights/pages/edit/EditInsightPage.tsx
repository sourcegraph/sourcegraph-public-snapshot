import classnames from 'classnames'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useContext, useMemo, useState } from 'react'
import { useHistory, Link } from 'react-router-dom'

import { PlatformContextProps } from '@sourcegraph/shared/src/platform/context'
import { Settings, SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { asError, isErrorLike } from '@sourcegraph/shared/src/util/errors'

import { AuthenticatedUser } from '../../../auth'
import { HeroPage } from '../../../components/HeroPage'
import { Page } from '../../../components/Page'
import { PageTitle } from '../../../components/PageTitle'
import { FORM_ERROR, SubmissionErrors } from '../../components/form/hooks/useForm'
import { InsightsApiContext } from '../../core/backend/api-provider'
import { addInsightToCascadeSetting, removeInsightFromSetting } from '../../core/jsonc-operation'
import { Insight, isLangStatsInsight, isSearchBasedInsight } from '../../core/types'

import { EditLangStatsInsight } from './components/EditLangStatsInsight'
import { EditSearchBasedInsight } from './components/EditSearchInsight'
import styles from './EditInsightPage.module.scss'

export interface EditInsightPageProps extends SettingsCascadeProps, PlatformContextProps<'updateSettings'> {
    /** Normalized insight id <type insight>.insight.<name of insight> */
    insightID: string

    /**
     * Authenticated user info, Used to decide where code insight will appears
     * in personal dashboard (private) or in organisation dashboard (public)
     * */
    authenticatedUser: Pick<AuthenticatedUser, 'id' | 'organizations' | 'username'>
}

interface ParsedInsightInfo {
    insight?: Insight | null
    originSubjectID?: string
}

export const EditInsightPage: React.FunctionComponent<EditInsightPageProps> = props => {
    const { insightID, settingsCascade, authenticatedUser, platformContext } = props
    const { getSubjectSettings, updateSubjectSettings } = useContext(InsightsApiContext)
    const history = useHistory()

    // We need to catch the settings only once during the first render otherwise
    // if we used useMemo then after we update the settings further in the submit
    // handler we will again try to find an insight that may no longer exist and
    // (if user changed visibility we remove insight first from previous subject)
    // show the wrong visual state.
    const [{ insight, originSubjectID }] = useState<ParsedInsightInfo>(() => {
        if (!authenticatedUser) {
            return {}
        }

        const subjects = settingsCascade.subjects
        const { id: userID } = authenticatedUser

        const subject = subjects?.find(({ settings }) => settings && !isErrorLike(settings) && !!settings[insightID])

        if (!subject?.settings || isErrorLike(subject.settings)) {
            return {}
        }

        // Form insight object from user/org settings to pass that info as
        // initial values for edit components
        const insight: Insight = {
            id: insightID,
            visibility: userID === subject.subject.id ? 'personal' : subject.subject.id,
            ...subject.settings[insightID],
        }

        return {
            insight,
            originSubjectID: subject.subject.id,
        }
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

    const handleSubmit = async (newInsight: Insight): Promise<SubmissionErrors> => {
        if (!insight || !originSubjectID || !authenticatedUser) {
            return
        }

        try {
            // Since insights live in user/org settings if visibility setting
            // has been changed we need remove previous (old) insight from previous
            // subject settings (user or org) and create new insight to new setting file.
            if (insight.visibility !== newInsight.visibility) {
                const settings = await getSubjectSettings(originSubjectID).toPromise()
                const editedSettings = removeInsightFromSetting(settings.contents, insight.id)

                await updateSubjectSettings(platformContext, originSubjectID, editedSettings).toPromise()
            }

            const { id: userID } = authenticatedUser

            const subjectID = newInsight.visibility === 'personal' ? userID : newInsight.visibility

            const settings = await getSubjectSettings(subjectID).toPromise()
            let settingsContent = settings.contents

            // Since id of insight is based on insight title if title was changed
            // we need remove old insight object from settings by insight old id
            if (insight.title !== newInsight.title) {
                settingsContent = removeInsightFromSetting(settingsContent, insight.id)
            }

            settingsContent = addInsightToCascadeSetting(settingsContent, newInsight)

            await updateSubjectSettings(platformContext, subjectID, settingsContent).toPromise()

            history.push('/insights')
        } catch (error) {
            return { [FORM_ERROR]: asError(error) }
        }

        return
    }

    if (!insight) {
        return (
            <HeroPage
                icon={MapSearchIcon}
                title="Oops, we couldn't find insight"
                subtitle={
                    <span>
                        We couldn't find that insight. Try to find the insight with ID:{' '}
                        <code className="badge badge-secondary">{insightID}</code> in your{' '}
                        {authenticatedUser ? (
                            <Link to={`/users/${authenticatedUser?.username}/settings`}>user or org settings</Link>
                        ) : (
                            <span>user or org settings</span>
                        )}
                    </span>
                }
            />
        )
    }

    const {
        organizations: { nodes: orgs },
    } = authenticatedUser

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
                    organizations={orgs}
                    onSubmit={handleSubmit}
                />
            )}

            {isLangStatsInsight(insight) && (
                <EditLangStatsInsight
                    insight={insight}
                    finalSettings={finalSettings}
                    organizations={orgs}
                    onSubmit={handleSubmit}
                />
            )}
        </Page>
    )
}
