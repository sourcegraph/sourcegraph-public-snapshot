import { FunctionComponent, useEffect } from 'react'
import { Redirect } from 'react-router-dom'

import { PageRoutes } from '../routes.constants'
import { useFeatureFlag } from '../featureFlags/useFeatureFlag'

export const GetStarted: FunctionComponent<IProps> = () => {
    const [enabled, status] = useFeatureFlag('ab_unified_registration')

    useEffect(() => {
        if (status !== 'loading' && !enabled) {
            window.location.href = 'https://about.sourcegraph.com/get-started'
        }
    }, [enabled, status])

    if (status === 'loading') {
        return null
    }

    if (enabled) {
        return <Redirect to={'/sign-up?returnTo=/deployment-options'} />
    }

    return null
}

export default GetStarted
