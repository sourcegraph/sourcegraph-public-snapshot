import React, { useMemo } from 'react'
import { useObservable } from '../../../shared/src/util/useObservable'
import { getViewsForContainer } from '../../../shared/src/api/client/services/viewService'
import { ContributableViewContainer } from '../../../shared/src/api/protocol'
import { ExtensionsControllerProps } from '../../../shared/src/extensions/controller'
import { ViewGrid, ViewGridProps } from '../repo/tree/ViewGrid'
import { InsightsIcon } from './icon'
import PlusIcon from 'mdi-react/PlusIcon'
import { Link } from '../../../shared/src/components/Link'
import GearIcon from 'mdi-react/GearIcon'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'

interface InsightsPageProps extends ExtensionsControllerProps, Omit<ViewGridProps, 'views'> {}
export const InsightsPage: React.FunctionComponent<InsightsPageProps> = props => {
    const views = useObservable(
        useMemo(
            () =>
                getViewsForContainer(
                    ContributableViewContainer.InsightsPage,
                    {},
                    props.extensionsController.services.view
                ),
            [props.extensionsController.services.view]
        )
    )
    return (
        <div className="container mt-3">
            <div className="d-flex align-items-center">
                <h1 className="flex-grow-1 text-nowrap">
                    <InsightsIcon className="icon-inline" /> Insights{' '}
                    <sup>
                        <span className="badge badge-primary text-uppercase">prototype</span>
                    </sup>
                </h1>
                {/* These buttons are just links until there is a proper configuration UI */}
                <Link to="/user/settings" className="btn btn-secondary mr-1">
                    <GearIcon className="icon-inline" /> Configure insights
                </Link>
                <Link to="/extensions?query=category:Insights" className="btn btn-secondary">
                    <PlusIcon className="icon-inline" /> Add more insights
                </Link>
            </div>
            {views === undefined ? (
                <div className="d-flex w-100">
                    <LoadingSpinner className="my-4" />
                </div>
            ) : (
                <ViewGrid {...props} views={views} />
            )}
        </div>
    )
}
