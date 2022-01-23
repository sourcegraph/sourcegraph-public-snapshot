import classNames from 'classnames'
import { Location, LocationDescriptorObject } from 'history'
import React, { useCallback, useMemo } from 'react'
import { useHistory, useLocation } from 'react-router'

import { isDefined } from '@sourcegraph/common'
import { SettingsCascadeProps } from '@sourcegraph/shared/src/settings/settings'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { BreadcrumbSetters } from '@sourcegraph/web/src/components/Breadcrumbs'

import {
    RepositoryForTreeFields,
    TreeEntryForTreeFields,
    TreeOrComponentPageResult,
    TreeOrComponentSourceLocationSetFields,
} from '../../../../graphql-operations'
import { ComponentActionPopoverButton } from '../../../../repo/actions/source-location-set-view-mode-action/SourceLocationSetViewModeAction'
import { CatalogPage, CatalogPage2 } from '../../components/catalog-area-header/CatalogPage'
import { CatalogRelations } from '../../pages/component/CatalogRelations'
import { CodeTab } from '../../pages/component/code/CodeTab'
import { UsageTab } from '../../pages/component/usage/UsageTab'
import { WhoKnowsTab } from '../../pages/component/who-knows/WhoKnowsTab'

import styles from './TreeOrComponent.module.scss'
import { TreeOrComponentHeader } from './TreeOrComponentHeader'

interface Props extends SettingsCascadeProps, TelemetryProps, BreadcrumbSetters {
    data: Extract<TreeOrComponentPageResult['node'], { __typename: 'Repository' }>
}

const tabContentClassName = classNames('flex-1 align-self-stretch', styles.tabContent)

export const TreeOrComponent: React.FunctionComponent<Props> = ({ data, useBreadcrumb, ...props }) => {
    const primaryComponent = data.primaryComponents.length > 0 ? data.primaryComponents[0] : null

    const treeOrComponentViewOptions = useTreeOrComponentViewOptions()

    const repository: RepositoryForTreeFields = data
    const tree: TreeEntryForTreeFields | null = data.commit?.tree ?? null

    const sourceLocationSet: TreeOrComponentSourceLocationSetFields | null =
        treeOrComponentViewOptions.treeOrComponentViewMode === 'auto' ? primaryComponent ?? tree : tree

    useBreadcrumb(
        useMemo(
            () =>
                primaryComponent
                    ? {
                          key: 'component',
                          className: 'flex-shrink-past-contents align-self-stretch',
                          element: (
                              <ComponentActionPopoverButton
                                  component={primaryComponent}
                                  {...treeOrComponentViewOptions}
                              />
                          ),
                          divider: <span className="mx-1" />,
                      }
                    : null,
            [primaryComponent, treeOrComponentViewOptions]
        )
    )

    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                {
                    path: ['', 'contributors', 'code-owners', 'commits', 'branches'],
                    text: 'Code',
                    content: (
                        <CodeTab
                            {...props}
                            {...treeOrComponentViewOptions}
                            repository={repository}
                            tree={tree}
                            component={primaryComponent}
                            sourceLocationSet={sourceLocationSet}
                            isTree={true}
                            useHash={true}
                            className={tabContentClassName}
                        />
                    ),
                },
                {
                    path: 'who-knows',
                    text: 'Who knows?',
                    content: (
                        <WhoKnowsTab
                            {...props}
                            sourceLocationSet={sourceLocationSet.id}
                            className={tabContentClassName}
                        />
                    ),
                },
                sourceLocationSet && sourceLocationSet.__typename === 'Component'
                    ? {
                          path: 'graph',
                          text: 'Graph',
                          content: (
                              <div className={classNames('p-3', tabContentClassName)}>
                                  <CatalogRelations
                                      component={sourceLocationSet.id}
                                      useURLForConnectionParams={true}
                                      className="mb-3"
                                  />
                              </div>
                          ),
                      }
                    : null,
                sourceLocationSet?.usage && {
                    path: 'usage',
                    text: 'Usage',
                    content: (
                        <UsageTab {...props} sourceLocationSet={sourceLocationSet.id} className={tabContentClassName} />
                    ),
                },
            ].filter(isDefined),
        [primaryComponent, props, repository, sourceLocationSet, tree, treeOrComponentViewOptions]
    )

    return (
        <CatalogPage2
            header={
                <TreeOrComponentHeader
                    repository={repository}
                    tree={tree}
                    primaryComponent={primaryComponent}
                    {...treeOrComponentViewOptions}
                />
            }
            tabs={tabs}
            useHash={true}
            tabsClassName={styles.tabs}
        />
    )
}

type TreeOrComponentViewMode = 'auto' | 'tree'

export interface TreeOrComponentViewOptionsProps {
    treeOrComponentViewMode: TreeOrComponentViewMode
    treeOrComponentViewModeURL: Record<TreeOrComponentViewMode, LocationDescriptorObject>
    setTreeOrComponentViewMode: (mode: TreeOrComponentViewMode) => void
}

export function useTreeOrComponentViewOptions(): TreeOrComponentViewOptionsProps {
    const location = useLocation()
    const history = useHistory()

    const treeOrComponentViewMode: TreeOrComponentViewMode = useMemo(
        () => (new URLSearchParams(location.search).get('as') === 'tree' ? 'tree' : 'auto'),
        [location.search]
    )
    const treeOrComponentViewModeURL = useMemo<TreeOrComponentViewOptionsProps['treeOrComponentViewModeURL']>(
        () => ({
            auto: makeTreeOrComponentViewURL(location, 'auto'),
            tree: makeTreeOrComponentViewURL(location, 'tree'),
        }),
        [location]
    )
    const setTreeOrComponentViewMode = useCallback<TreeOrComponentViewOptionsProps['setTreeOrComponentViewMode']>(
        mode => history.push(makeTreeOrComponentViewURL(location, mode)),
        [history, location]
    )

    return useMemo(() => ({ treeOrComponentViewMode, treeOrComponentViewModeURL, setTreeOrComponentViewMode }), [
        setTreeOrComponentViewMode,
        treeOrComponentViewMode,
        treeOrComponentViewModeURL,
    ])
}

function makeTreeOrComponentViewURL(location: Location, mode: TreeOrComponentViewMode): LocationDescriptorObject {
    const search = new URLSearchParams(location.search)
    if (mode === 'tree') {
        search.set('as', mode)
    } else {
        search.delete('as')
    }

    return { ...location, search: search.toString() }
}
