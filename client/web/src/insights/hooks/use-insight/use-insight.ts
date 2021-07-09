
import { useMemo } from 'react';

import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings';
import { isErrorLike } from '@sourcegraph/shared/src/util/errors';

import { Settings } from '../../../schema/settings.schema';
import { useInsightDashboards } from '../use-insight-dashboards/use-insight-dashboards';
import { Insight, InsightSettings } from '../../core/types';

export interface UseInsightProps extends SettingsCascadeProps<Settings> {
    insightId: string
}

/**
 * Returns parsed insight from the settings cascade configuration.
 */
export function useInsight(props: UseInsightProps): Insight | null {
    const { insightId, settingsCascade } = props;

    const insightDashboards = useInsightDashboards({ insightId, settingsCascade })

    return useMemo(() => {
        const subjects = settingsCascade.subjects
        const subject = subjects?.find(({ settings }) => settings && !isErrorLike(settings) && !!settings[insightId])

        if (!subject?.settings || isErrorLike(subject.settings)) {
            return null
        }

        // Form insight object from user/org settings to pass that info as
        // initial values for edit components
        const insight: Insight = {
            id: insightId,
            dashboards: insightDashboards,
            storeSubject: subject,
            ...subject.settings[insightId] as InsightSettings,
        }

        return insight
    }, [insightDashboards, insightId, settingsCascade.subjects])
}
