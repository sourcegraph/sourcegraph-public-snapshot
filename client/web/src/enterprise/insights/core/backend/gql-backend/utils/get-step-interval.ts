import type { Duration } from 'date-fns'

import { isDefined } from '@sourcegraph/common'
import { TimeIntervalStepUnit } from '@sourcegraph/shared/src/graphql-operations'

/**
 * Returns tuple with gql model time interval unit and value. Used to convert FE model
 * insight data series time step to GQL time interval model.
 */
export function getStepInterval(step: Duration): [TimeIntervalStepUnit, number] {
    const castUnits = (Object.keys(step) as (keyof Duration)[])
        .map<[TimeIntervalStepUnit, number] | null>(key => {
            switch (key) {
                case 'hours': {
                    return [TimeIntervalStepUnit.HOUR, step[key] ?? 0]
                }
                case 'days': {
                    return [TimeIntervalStepUnit.DAY, step[key] ?? 0]
                }
                case 'weeks': {
                    return [TimeIntervalStepUnit.WEEK, step[key] ?? 0]
                }
                case 'months': {
                    return [TimeIntervalStepUnit.MONTH, step[key] ?? 0]
                }
                case 'years': {
                    return [TimeIntervalStepUnit.YEAR, step[key] ?? 0]
                }
            }

            return null
        })
        .filter(isDefined)

    if (castUnits.length === 0) {
        throw new Error('Wrong time step format')
    }

    // Return first valid match
    return castUnits[0]
}
