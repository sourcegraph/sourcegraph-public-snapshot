import React from 'react'

import { Routes, Route } from 'react-router-dom-v5-compat'

import { Page } from '../../components/Page'
import { useFeatureFlag } from '../../featureFlags/useFeatureFlag'

import { CodyPage } from './CodyPage'

/**
 * The global Cody area.
 *
 * For Sourcegraph team members only. For instructions, see
 * https://docs.google.com/document/d/1u7HYPmJFtDANtBgczzmAR0BmhM86drwDXCqx-F2jTEE/edit#.
 */
export const GlobalCodyArea: React.FunctionComponent = ({ ...outerProps }) => {
    const [codyEnabled] = useFeatureFlag('cody')
    if (!codyEnabled) {
        return <Page>Cody is not enabled.</Page>
    }

    return (
        <div className="w-100">
            <Page>
                <Routes>
                    <Route path="" element={<CodyPage {...outerProps} />} />
                </Routes>
            </Page>
        </div>
    )
}
