export enum GroupResultsPing {
    // Aggregation chart events
    ChartBarClick = 'GroupResultsChartBarClick',
    ChartBarHover = 'GroupResultsChartBarHover',

    // Aggregation mode events
    ModeClick = 'GroupAggregationModeClicked',
    ModeDisabledHover = 'GroupAggregationModeDisabledHover',

    // Other UI
    CollapseSidebarSection = 'GroupResultsCollapseSection',
    ExpandSidebarSection = 'GroupResultsOpenSection',
    ExpandFullViewPanel = 'GroupResultsExpandedViewOpen',
    CollapseFullViewPanel = 'GroupResultsExpandedViewCollapse',
    InfoIconHover = 'GroupResultsInfoIconHover',

    // Proactive
    ProactiveLimitHit = 'ProactiveLimitHit',
    ProactiveLimitSuccess = 'ProactiveLimitSuccess',
    ExplicitLimitHit = 'ExplicitLimitHit',
    ExplicitLimitSuccess = 'ExplicitLimitSuccess',
}
