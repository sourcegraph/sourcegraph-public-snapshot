import Axe from 'axe-core'

import { formatRuleViolations2, RuleViolation } from './formatAxeViolations'

export async function runtimeAccessibilityAudit(): Promise<RuleViolation[]> {
    const rootElement = document.querySelector('#root')

    if (!rootElement) {
        throw new Error('No root element found')
    }

    const { violations } = await Axe.run(rootElement)
    const formattedViolations = formatRuleViolations2(violations)
    return formattedViolations
}
