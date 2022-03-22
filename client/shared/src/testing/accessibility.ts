import { AxePuppeteer } from '@axe-core/puppeteer'
import type { Result, NodeResult, RunOptions } from 'axe-core'
import { Page } from 'puppeteer'

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
 * Filters any nodes that should be explicitly ignored from our Axe violations.
 */
const filterIgnoredNodes = (nodes: NodeResult[]): NodeResult[] =>
    nodes.filter(({ target }) => !target.includes('a11y-ignore'))

/**
 * Takes a list of Axe violation and formats them into readable strings.
 */
const formatRuleViolations = (violations: Result[]): string[] =>
    violations
        .map(violation => ({
            ...violation,
            nodes: filterIgnoredNodes(violation.nodes),
        }))
        .filter(({ nodes }) => nodes.length > 0)
        .map(
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

export async function accessibilityAudit(page: Page, config: AccessibilityAuditConfiguration = {}): Promise<void> {
    const { options, mode = 'fail' } = config
    const axe = new AxePuppeteer(page)

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

        console.warn(errorMessage)
    }
}
