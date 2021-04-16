/**
 * Returns insight type according to viewID and our naming convention
 * [extensionName].[name/id of code insight].[place = Home | Insight page | Directory]
 */
export const getInsightTypeByViewId = (viewID: string): string => viewID.split('.')[0]
