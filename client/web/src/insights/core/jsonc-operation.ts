import * as jsonc from '@sqs/jsonc-parser'

import { Insight } from './types'

export const defaultFormattingOptions: jsonc.FormattingOptions = {
    eol: '\n',
    insertSpaces: true,
    tabSize: 2,
}

export const addInsightToCascadeSetting = (settings: string, insight: Insight): string => {
    const { id, visibility, ...originInsight } = insight

    const edits = jsonc.modify(settings, [id], originInsight, { formattingOptions: defaultFormattingOptions })

    return jsonc.applyEdits(settings, edits)
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
