import H from 'history'
import MapSearchIcon from 'mdi-react/MapSearchIcon'
import React, { useMemo, useState } from 'react'
import { Route, RouteComponentProps, Switch } from 'react-router'
import { ExtensionsControllerProps } from '../../../../../../shared/src/extensions/controller'
import * as GQL from '../../../../../../shared/src/graphql/schema'
import { PlatformContextProps } from '../../../../../../shared/src/platform/context'
import { isDefined } from '../../../../../../shared/src/util/types'
import { BreadcrumbItem, Breadcrumbs } from '../../../../components/breadcrumbs/Breadcrumbs'
import { HeroPage } from '../../../../components/HeroPage'
import { RuleArea } from '../detail/RuleArea'
import { RulesNewPage } from '../new/RulesNewPage'
import { RuleScope } from '../types'
import { RulesListPage } from './list/RulesListPage'

export interface RulesAreaContext extends ExtensionsControllerProps, PlatformContextProps {
    /** The rule scope. */
    scope: RuleScope

    /** The URL to the rules area. */
    rulesURL: string

    setBreadcrumbItem: (breadcrumbItem: BreadcrumbItem | undefined) => void

    location: H.Location
    history: H.History
    authenticatedUser: GQL.IUser | null
    isLightTheme: boolean
}

interface Props extends Pick<RulesAreaContext, Exclude<keyof RulesAreaContext, 'rulesURL'>>, RouteComponentProps<{}> {}

/**
 * The rules area for a particular scope.
 */
export const RulesArea: React.FunctionComponent<Props> = ({ scope, match, ...props }) => {
    const [breadcrumbItem, setBreadcrumbItem] = useState<BreadcrumbItem>()

    const context: RulesAreaContext = {
        ...props,
        scope,
        rulesURL: match.url,
        setBreadcrumbItem,
    }
    const newRuleURL = `${context.rulesURL}/new`

    const breadcrumbItems: BreadcrumbItem[] = useMemo(
        () =>
            [{ text: scope.name, to: scope.url }, { text: 'Rules', to: context.rulesURL }, breadcrumbItem].filter(
                isDefined
            ),
        [breadcrumbItem, context.rulesURL, scope.name, scope.url]
    )

    return (
        <div className="container">
            <Breadcrumbs items={breadcrumbItems} className="my-4" />
            <Switch>
                <Route path={match.url} exact={true}>
                    <RulesListPage {...context} newRuleURL={newRuleURL} />
                </Route>
                <Route path={newRuleURL} exact={true}>
                    <RulesNewPage {...context} />
                </Route>
                <Route
                    path={`${match.url}/:ruleID`}
                    // tslint:disable-next-line:jsx-no-lambda
                    render={(routeComponentProps: RouteComponentProps<{ ruleID: string }>) => (
                        <RuleArea {...context} ruleID={routeComponentProps.match.params.ruleID} />
                    )}
                />
                <Route>
                    <HeroPage
                        icon={MapSearchIcon}
                        title="404: Not Found"
                        subtitle="Sorry, the requested page was not found."
                    />
                </Route>
            </Switch>
        </div>
    )
}
