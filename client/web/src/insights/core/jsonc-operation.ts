import * as jsonc from '@sqs/jsonc-parser'

import { Insight, isLangStatsInsight, isSearchBasedInsight } from './types'

export const defaultFormattingOptions: jsonc.FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

const getExtensionNameByInsight = (insight: Insight): string => {
    if (isSearchBasedInsight(insight)) {
        return 'sourcegraph/search-insights'
    }

    if (isLangStatsInsight(insight)) {
        return 'sourcegraph/code-stats-insights'
    }

    return ''
}

export const addInsightToCascadeSetting = (settings: string, insight: Insight): string => {
    const { id, visibility, ...originInsight } = insight

    const extensionName = getExtensionNameByInsight(insight)
    // Turn on extension if user in creation code insight.
    const addingExtensionKeyEdits = jsonc.modify(settings, ['extensions', extensionName], true, {
        formattingOptions: defaultFormattingOptions,
    })
    const addingInsightEdits = jsonc.modify(settings, [id], originInsight, {
        formattingOptions: defaultFormattingOptions,
    })

    return jsonc.applyEdits(settings, [...addingExtensionKeyEdits, ...addingInsightEdits])
}

export const removeInsightFromSetting = (settings: string, insightID: string): string => {
    const edits = jsonc.modify(
        settings,
        // According to our naming convention <type>.insight.<name>
        [`${insightID}`],
        undefined,
        { formattingOptions: defaultFormattingOptions }
    )

    return jsonc.applyEdits(settings, edits)
}
