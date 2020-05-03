import React from 'react'
import { ErrorLike, isErrorLike } from '../../../../shared/src/util/errors'
import { View } from 'sourcegraph'
import classNames from 'classnames'
import { LoadingSpinner } from '@sourcegraph/react-loading-spinner'
import { ErrorAlert } from '../../components/alerts'
import { ViewContent, ViewContentProps } from '../../views/ViewContent'
import * as H from 'history'

interface ViewGridProps extends Omit<ViewContentProps, 'viewContent'> {
    views: (ErrorLike | View | undefined)[]
    className?: string
    history: H.History
}

export const ViewGrid: React.FunctionComponent<ViewGridProps> = ({ views, className, ...props }) => (
    <div className={classNames(className, 'view-grid')}>
        {views.map((view, i) => (
            <div key={i} className="card">
                {view === undefined ? (
                    <div className="card-body d-flex flex-column align-items-center p-5">
                        <LoadingSpinner /> Loading code insight
                    </div>
                ) : isErrorLike(view) ? (
                    <ErrorAlert className="m-0" error={view} history={props.history} />
                ) : (
                    <div className="card-body">
                        <h3 className="view-grid__view-title">{view.title}</h3>
                        <ViewContent {...props} viewContent={view.content} />
                    </div>
                )}
            </div>
        ))}
    </div>
)
