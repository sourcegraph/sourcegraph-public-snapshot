import React, { useMemo } from 'react'
import { StatusBarItemWithKey } from '../../../../shared/src/api/client/api/codeEditor'
import { AbsoluteRepoFile } from '../../../../shared/src/util/url'
import { useObservable } from '../../../../shared/src/util/useObservable'
import classNames from 'classnames'
import { TextDocumentIdentifier } from '../../../../shared/src/api/client/types/textDocument'
import { Observable, timer } from 'rxjs'
import { useCarousel } from '../../components/useCarousel'
import MenuLeftIcon from 'mdi-react/MenuLeftIcon'
import MenuRightIcon from 'mdi-react/MenuRightIcon'
import { mapTo } from 'rxjs/operators'
import { Link } from 'react-router-dom'

interface StatusBarProps extends Omit<AbsoluteRepoFile, 'revision'> {
    getStatusBarItems: (parameters: TextDocumentIdentifier) => Observable<StatusBarItemWithKey[] | null>
}

export const StatusBar: React.FunctionComponent<StatusBarProps> = ({
    getStatusBarItems,
    repoName,
    filePath,
    commitID,
}) => {
    const uri = `git://${repoName}?${commitID}#${filePath}`

    const statusBarItems = useObservable(
        useMemo(
            () =>
                getStatusBarItems({
                    uri,
                }),
            [getStatusBarItems, uri]
        )
    )

    // Wait 3 seconds to show "no information from extensions avaiable"
    // to avoid UI jitter during initial extension activation.
    //
    // Restart timer whenever uri changes, since new extensions could be activated,
    // or could be fetching new data, so we include an unnecessary dependency.
    // eslint-disable-next-line react-hooks/exhaustive-deps
    const hadTimeToLoad = useObservable(useMemo(() => timer(3000).pipe(mapTo(true)), [uri]))

    const {
        carouselReference,
        canScrollNegative,
        canScrollPositive,
        onNegativeClicked,
        onPositiveClicked,
    } = useCarousel({ direction: 'leftToRight' })

    return (
        <div className="status-bar w-100 border-top d-flex">
            {canScrollNegative && (
                <button type="button" className="btn btn-link status-bar__scroll border-0" onClick={onNegativeClicked}>
                    <MenuLeftIcon className="icon-inline" />
                </button>
            )}
            <div className="status-bar__items d-flex px-2" ref={carouselReference}>
                {statusBarItems
                    ? statusBarItems.map(statusBarItem => (
                          <StatusBarItem key={statusBarItem.key} statusBarItem={statusBarItem} />
                      ))
                    : hadTimeToLoad && (
                          <StatusBarItem
                              key="none-found"
                              statusBarItem={{
                                  key: 'none-found',
                                  text:
                                      'No information from extensions available. Find extensions in the Sourcegraph extension registry',
                              }}
                              component={
                                  <small className="text-muted">
                                      No information from extensions available.{' '}
                                      <Link to="/extensions">
                                          Find extensions in the Sourcegraph extension registry
                                      </Link>
                                  </small>
                              }
                          />
                      )}
            </div>
            {canScrollPositive && (
                <button type="button" className="btn btn-link status-bar__scroll border-0" onClick={onPositiveClicked}>
                    <MenuRightIcon className="icon-inline" />
                </button>
            )}
        </div>
    )
}

const StatusBarItem: React.FunctionComponent<{
    statusBarItem: StatusBarItemWithKey
    className?: string
    component?: JSX.Element
}> = ({ statusBarItem, className = 'status-bar', component }) => {
    const handleCommand = () => {}

    return (
        <div
            className={classNames(
                `${className}__item h-100 d-flex align-items-center px-1`,
                statusBarItem.tooltip && `${className}__item--tooltipped`
            )}
            data-tooltip={statusBarItem.tooltip}
            onClick={handleCommand}
        >
            {component || <small className={`${className}__text`}>{statusBarItem.text}</small>}
        </div>
    )
}
