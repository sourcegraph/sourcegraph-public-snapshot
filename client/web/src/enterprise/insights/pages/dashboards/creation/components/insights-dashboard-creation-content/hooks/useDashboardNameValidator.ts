import { camelCase } from 'lodash'
import { useMemo } from 'react'

import { InsightDashboard } from '../../../../../../../../schema/settings.schema'
import { Validator } from '../../../../../../components/form/hooks/useField'
import { composeValidators, createRequiredValidator } from '../../../../../../components/form/validators'

interface useDashboardNameValidatorProps {
    settings: {
        [k: string]: InsightDashboard
    }
}

/**
 * Dashboard's name validator hook.
 * Dashboard's name is required and must be unique for all insights dashboards.
 */
export function useDashboardNameValidator(props: useDashboardNameValidatorProps): Validator<string> {
    const { settings } = props

    return useMemo(() => {
        const existingDashboardNames = new Set(Object.keys(settings))

        return composeValidators<string>(createRequiredValidator('Name is a required field.'), value =>
            // According to our name dashboard convention, that dashboard id equals camel-cased name.
            existingDashboardNames.has(camelCase(value))
                ? 'A dashboard with this name already exists. Please set a different name for the new dashboard.'
                : undefined
        )
    }, [settings])
}
