import { SettingsSiteSubject } from '@sourcegraph/shared/src/settings/settings'

/**
 * Returns tooltip text for global subject visibility option.
 */
export function getGlobalSubjectTooltipText(globalSubject: SettingsSiteSubject | undefined): string | undefined {
    if (!globalSubject) {
        return
    }

    return !globalSubject.viewerCanAdminister ? 'Only site admins can create global dashboards' : undefined
}
