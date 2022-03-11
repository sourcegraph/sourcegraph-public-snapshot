import copy from 'copy-to-clipboard'
import ContentCopyIcon from 'mdi-react/ContentCopyIcon'
import React, { MouseEvent, useContext, useState } from 'react'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import {
    Button,
    Card,
    CardBody,
    CardText,
    CardTitle,
    Tab,
    TabList,
    TabPanel,
    TabPanels,
    Tabs,
    TooltipController,
    Icon,
    Link,
} from '@sourcegraph/wildcard'

import { InsightType } from '../../../../../core/types'
import { encodeCaptureInsightURL } from '../../../../insights/creation/capture-group'
import { encodeSearchInsightUrl } from '../../../../insights/creation/search-insight'
import {
    CodeInsightsLandingPageContext,
    CodeInsightsLandingPageType,
    useLogEventName,
} from '../../../CodeInsightsLandingPageContext'
import { CodeInsightsQueryBlock } from '../code-insights-query-block/CodeInsightsQueryBlock'

import styles from './CodeInsightsTemplates.module.scss'
import { Template, TEMPLATE_SECTIONS } from './constants'

function getTemplateURL(template: Template): string {
    switch (template.type) {
        case InsightType.CaptureGroup:
            return `/insights/create/capture-group?${encodeCaptureInsightURL(template.templateValues)}`
        case InsightType.SearchBased:
            return `/insights/create/search?${encodeSearchInsightUrl(template.templateValues)}`
    }
}

interface CodeInsightsTemplates extends TelemetryProps, React.HTMLAttributes<HTMLElement> {}

export const CodeInsightsTemplates: React.FunctionComponent<CodeInsightsTemplates> = props => {
    const { telemetryService, ...otherProps } = props
    const tabChangePingName = useLogEventName('InsightsGetStartedTabClick')

    const handleTabChange = (index: number): void => {
        const template = TEMPLATE_SECTIONS[index]

        telemetryService.log(tabChangePingName, { tabName: template.title }, { tabName: template.title })
    }

    return (
        <section {...otherProps}>
            <h2 id="code-insights-templates">Templates</h2>
            <p className="text-muted">
                Some of the most popular{' '}
                <Link to="/help/code_insights/references/common_use_cases" rel="noopener noreferrer" target="_blank">
                    use cases
                </Link>
                .
            </p>

            <Tabs size="medium" className="mt-3" onChange={handleTabChange}>
                <TabList wrapperClassName={styles.tabList}>
                    {TEMPLATE_SECTIONS.map(section => (
                        <Tab key={section.title}>{section.title}</Tab>
                    ))}
                </TabList>
                <TabPanels>
                    {TEMPLATE_SECTIONS.map(section => (
                        <TemplatesPanel
                            key={section.title}
                            sectionTitle={section.title}
                            templates={section.templates}
                            telemetryService={telemetryService}
                        />
                    ))}
                </TabPanels>
            </Tabs>
        </section>
    )
}

interface TemplatesPanelProps extends TelemetryProps {
    sectionTitle: string
    templates: Template[]
}

const TemplatesPanel: React.FunctionComponent<TemplatesPanelProps> = props => {
    const { templates, sectionTitle, telemetryService } = props
    const [allVisible, setAllVisible] = useState(false)
    const tabMoreClickPingName = useLogEventName('InsightsGetStartedTabMoreClick')

    const maxNumberOfCards = allVisible ? templates.length : 4
    const hasMoreLessButton = templates.length > 4

    const handleShowMoreButtonClick = (): void => {
        if (!allVisible) {
            telemetryService.log(tabMoreClickPingName, { tabName: sectionTitle }, { tabName: sectionTitle })
        }

        setAllVisible(!allVisible)
    }

    return (
        <TabPanel className={styles.cards}>
            {templates.slice(0, maxNumberOfCards).map(template => (
                <TemplateCard key={template.title} template={template} telemetryService={telemetryService} />
            ))}

            {hasMoreLessButton && (
                <Button
                    variant="secondary"
                    outline={true}
                    className={styles.cardsFooterButton}
                    onClick={handleShowMoreButtonClick}
                >
                    {allVisible ? 'Show less' : 'Show more'}
                </Button>
            )}
        </TabPanel>
    )
}

interface TemplateCardProps extends TelemetryProps {
    template: Template
}

const TemplateCard: React.FunctionComponent<TemplateCardProps> = props => {
    const { template, telemetryService } = props
    const { mode } = useContext(CodeInsightsLandingPageContext)

    const series =
        template.type === InsightType.SearchBased
            ? template.templateValues.series ?? []
            : [{ query: template.templateValues.groupSearchQuery }]

    const handleUseTemplateLinkClick = (): void => {
        telemetryService.log('InsightGetStartedTemplateClick')
    }

    return (
        <Card as={CardBody} className={styles.card}>
            <CardTitle>{template.title}</CardTitle>
            <CardText>{template.description}.</CardText>

            <div className={styles.queries}>
                {series.map(
                    line =>
                        line.query && (
                            <QueryPanel key={line.query} query={line.query} telemetryService={telemetryService} />
                        )
                )}
            </div>

            {mode === CodeInsightsLandingPageType.InProduct && (
                <Button
                    as={Link}
                    to={getTemplateURL(template)}
                    variant="secondary"
                    outline={true}
                    className="mr-auto"
                    onClick={handleUseTemplateLinkClick}
                >
                    Use this template
                </Button>
            )}
        </Card>
    )
}

interface QueryPanelProps extends TelemetryProps {
    query: string
}

const copyTooltip = 'Copy query'
const copyCompletedTooltip = 'Copied!'

const QueryPanel: React.FunctionComponent<QueryPanelProps> = props => {
    const { query, telemetryService } = props

    const templateCopyClickEvenName = useLogEventName('InsightGetStartedTemplateCopyClick')
    const [currentCopyTooltip, setCurrentCopyTooltip] = useState(copyTooltip)

    const onCopy = (event: MouseEvent<HTMLButtonElement>): void => {
        copy(query)
        setCurrentCopyTooltip(copyCompletedTooltip)
        setTimeout(() => setCurrentCopyTooltip(copyTooltip), 1000)

        requestAnimationFrame(() => {
            TooltipController.forceUpdate()
        })

        event.preventDefault()
        telemetryService.log(templateCopyClickEvenName)
    }

    return (
        <CodeInsightsQueryBlock className={styles.query}>
            <SyntaxHighlightedSearchQuery query={query} />
            <Button
                className={styles.copyButton}
                onClick={onCopy}
                data-tooltip={currentCopyTooltip}
                data-placement="top"
                aria-label="Copy Docker command to clipboard"
                variant="icon"
            >
                <Icon as={ContentCopyIcon} />
            </Button>
        </CodeInsightsQueryBlock>
    )
}
