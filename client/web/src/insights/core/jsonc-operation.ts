import * as jsonc from '@sqs/jsonc-parser'

import { Insight, isLangStatsInsight, isSearchBasedInsight } from './types'

export const defaultFormattingOptions: jsonc.FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

/**
 * Simplified jsonc API method to modify jsonc object.
 *
 * @param originContent Origin content (settings)
 * @param path - path to the field which will be modified
 * @param value - new value for modify field
 */
export const modify = (originContent: string, path: jsonc.JSONPath, value: unknown): string => {
    const addingExtensionKeyEdits = jsonc.modify(originContent, path, value, {
        formattingOptions: defaultFormattingOptions,
    })

    return jsonc.applyEdits(originContent, addingExtensionKeyEdits)
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
    const settingsWithExtension = modify(settings, ['extensions', extensionName], true)

    // Add insight to the user settings
    return modify(settingsWithExtension, [id], originInsight)
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
