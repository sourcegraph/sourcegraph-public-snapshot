import React, { useState } from 'react'
import { Link, LinkProps } from 'react-router-dom'

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
} from '@sourcegraph/wildcard'

import { InsightType } from '../../../../core/types'
import { encodeCaptureInsightURL } from '../../../insights/creation/capture-group'
import { encodeSearchInsightUrl } from '../../../insights/creation/search-insight'

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

export const CodeInsightsTemplates: React.FunctionComponent<React.HTMLAttributes<HTMLElement>> = props => (
    <section {...props}>
        <h2>Templates</h2>
        <p className="text-muted">
            Some of the most popular{' '}
            <a href="/help/code_insights/references/common_use_cases" rel="noopener noreferrer" target="_blank">
                use cases
            </a>
            , collected from our beta customers.
        </p>

        <Tabs size="medium" className="mt-3">
            <TabList>
                {TEMPLATE_SECTIONS.map(section => (
                    <Tab key={section.title}>{section.title}</Tab>
                ))}
            </TabList>
            <TabPanels>
                {TEMPLATE_SECTIONS.map(section => (
                    <TemplatesPanel key={section.title} templates={section.templates} />
                ))}
            </TabPanels>
        </Tabs>
    </section>
)

interface TemplatesPanelProps {
    templates: Template[]
}

const TemplatesPanel: React.FunctionComponent<TemplatesPanelProps> = props => {
    const { templates } = props
    const [allVisible, setAllVisible] = useState(false)

    const maxNumberOfCards = allVisible ? templates.length : 4
    const hasMoreLessButton = templates.length > 4

    return (
        <TabPanel className={styles.cards}>
            {templates.slice(0, maxNumberOfCards).map(template => (
                <Card key={template.title} as={TemplateCardBody} to={getTemplateURL(template)} className={styles.card}>
                    <CardTitle>{template.title}</CardTitle>
                    <CardText className="flex-grow-1">{template.description}</CardText>
                    <Button variant="secondary" outline={true} className="mr-auto">
                        Use this template
                    </Button>
                </Card>
            ))}

            {hasMoreLessButton && (
                <Button
                    variant="secondary"
                    outline={true}
                    className={styles.cardsFooterButton}
                    onClick={() => setAllVisible(!allVisible)}
                >
                    {allVisible ? 'Show less' : 'Show all'}
                </Button>
            )}
        </TabPanel>
    )
}

interface TemplateCardBodyProps extends LinkProps {}

export const TemplateCardBody: React.FunctionComponent<TemplateCardBodyProps> = props => (
    <CardBody as={Link} {...props} />
)
