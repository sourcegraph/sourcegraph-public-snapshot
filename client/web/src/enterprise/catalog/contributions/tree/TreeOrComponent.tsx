import React, { useMemo } from 'react'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { isDefined } from '@sourcegraph/shared/src/util/types'

import { TreeOrComponentPageResult } from '../../../../graphql-operations'
import { CatalogPage, CatalogPage2 } from '../../components/catalog-area-header/CatalogPage'
import { CodeTab } from '../../pages/component/code/CodeTab'
import { TAB_CONTENT_CLASS_NAME } from '../../pages/component/ComponentDetailContent'
import { OverviewTab } from '../../pages/component/overview/OverviewTab'
import { RelationsTab } from '../../pages/component/RelationsTab'
import { UsageTab } from '../../pages/component/UsageTab'

import { TreeOrComponentHeader } from './TreeOrComponentHeader'

interface Props extends TelemetryProps {
    data: Extract<TreeOrComponentPageResult['node'], { __typename: 'Repository' }>
}

export const TreeOrComponent: React.FunctionComponent<Props> = ({ data, telemetryService, ...props }) => {
    const primaryComponent = data.primaryComponents.length > 0 ? data.primaryComponents[0] : null

    const tabs = useMemo<React.ComponentProps<typeof CatalogPage>['tabs']>(
        () =>
            [
                primaryComponent && {
                    path: ['', 'who-knows'],
                    exact: true,
                    text: 'Overview',
                    content: <OverviewTab {...props} component={primaryComponent} className={TAB_CONTENT_CLASS_NAME} />,
                },

                primaryComponent && {
                    path: 'code',
                    text: 'Code',
                    content: <CodeTab {...props} component={primaryComponent} className={TAB_CONTENT_CLASS_NAME} />,
                },
                primaryComponent && {
                    path: 'graph',
                    text: 'Graph',
                    content: (
                        <RelationsTab {...props} component={primaryComponent} className={TAB_CONTENT_CLASS_NAME} />
                    ),
                },
                primaryComponent &&
                    primaryComponent.usage && {
                        path: 'usage',
                        text: 'Usage',
                        content: (
                            <UsageTab {...props} component={primaryComponent} className={TAB_CONTENT_CLASS_NAME} />
                        ),
                    },
            ].filter(isDefined),
        [primaryComponent, props]
    )

    return primaryComponent ? (
        <>
            <CatalogPage2
                header={<TreeOrComponentHeader primaryComponent={primaryComponent} />}
                tabs={tabs}
                useHash={true}
            />
        </>
    ) : (
        <p>No primary component</p>
    )
}
