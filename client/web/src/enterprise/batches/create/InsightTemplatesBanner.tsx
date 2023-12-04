import React from 'react'

import { TemplateBanner } from './TemplateBanner'

interface InsightTemplatesBannerProps {
    insightTitle: string
    type: 'create' | 'edit'
    className?: string
}

export const InsightTemplatesBanner: React.FunctionComponent<React.PropsWithChildren<InsightTemplatesBannerProps>> = ({
    insightTitle,
    type,
    className,
}) => {
    const [heading, description]: [React.ReactNode, React.ReactNode] =
        type === 'create'
            ? [
                  'You are creating a batch change from a code insight',
                  <>
                      Let Sourcegraph help you with <strong>{insightTitle}</strong> by preparing a relevant{' '}
                      <strong>batch change</strong>.
                  </>,
              ]
            : [
                  `Start from template for the ${insightTitle}`,
                  `Sourcegraph pre-selected a batch spec for the batch change started from ${insightTitle}.`,
              ]

    return <TemplateBanner heading={heading} description={description} className={className} />
}
