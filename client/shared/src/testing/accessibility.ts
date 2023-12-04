import { AxePuppeteer } from '@axe-core/puppeteer'
import type { Result, NodeResult, RunOptions } from 'axe-core'
import type { Page } from 'puppeteer'

import { logger } from '@sourcegraph/common'

/**
 * Takes a list of Axe violation nodes and formats them into a readable string.
 */
const formatViolationProblems = (nodes: NodeResult[]): string =>
    nodes
        .map(
            ({ failureSummary = '', html }) => `
${html}:
${failureSummary}
            `
        )
        .join('')

/**
 * Takes a list of Axe violation and formats them into readable strings.
 */
const formatRuleViolations = (violations: Result[]): string[] =>
    violations.map(
        ({ id, help, helpUrl, nodes }) => `
Rule: "${id}" (${help})
Further information: ${helpUrl}
How to manually audit: https://docs.sourcegraph.com/dev/background-information/web/accessibility/how-to-audit#auditing-a-user-journey
Required changes: ${formatViolationProblems(nodes)}
`
    )

interface AccessibilityAuditConfiguration {
    options?: RunOptions
    mode?: 'fail' | 'warn'
}

/**
 * Use this `CSS` class constant to ignore an element in an accessibility audit.
 */
export const ACCESSIBILITY_AUDIT_IGNORE_CLASS = '.a11y-ignore'

/**
 * Runs an accessibility audit for the current page.
 *
 * Will error with a list of violations if any are found.
 *
 * See further documentation: https://docs.sourcegraph.com/dev/how-to/testing#accessibility-tests
 */
export async function accessibilityAudit(page: Page, config: AccessibilityAuditConfiguration = {}): Promise<void> {
    const { options, mode = 'fail' } = config
    const axe = new AxePuppeteer(page)
        .exclude(ACCESSIBILITY_AUDIT_IGNORE_CLASS)
        // https://github.com/microsoft/monaco-editor/issues/2448
        .exclude('.monaco-status')
        /*
         * TODO: Design review on some CodeMirror query input features to choose
         * a color that fulfill contrast requirements:
         * https://github.com/sourcegraph/sourcegraph/issues/36534
         * Additionally role="combobox" cannot be used together with aria-multiline, which
         * CodeMirror sets by default.
         */
        .exclude('.cm-content')

    if (options) {
        axe.options(options)
    }

    const { violations } = await axe.analyze()
    const formattedViolations = formatRuleViolations(violations)

    if (formattedViolations.length > 0) {
        const errorMessage = `Accessibility audit failed, ${
            formattedViolations.length
        } rule violations found:\n${formattedViolations.join('\n')}`

        if (mode === 'fail') {
            throw new Error(errorMessage)
        }

        logger.warn(errorMessage)
    }
}
