import React, { useLayoutEffect, useState } from 'react'

import classNames from 'classnames'
import { throttle } from 'lodash'

import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { Button, Card, CardBody, Link, Typography } from '@sourcegraph/wildcard'

import { CodeInsightExample } from '../../../getting-started/components/code-insights-examples/CodeInsightsExamples'

import { CodeInsightsExamplesSlider } from './code-insights-examples-slider/CodeInsightsExamplesSlider'
import { EXAMPLES } from './examples'

import styles from './CodeInsightsExamplesPicker.module.scss'

interface CodeInsightsExamplesPickerProps extends TelemetryProps {}

export const CodeInsightsExamplesPicker: React.FunctionComponent<
    React.PropsWithChildren<CodeInsightsExamplesPickerProps>
> = ({ telemetryService }) => {
    const [activeExampleIndex, setActiveExampleIndex] = useState(0)
    const [windowSize, setWindowSize] = useState(0)

    useLayoutEffect(() => {
        setWindowSize(window.innerWidth)
        const handleWindowResize = throttle(() => setWindowSize(window.innerWidth), 200)

        window.addEventListener('resize', handleWindowResize)

        return () => window.removeEventListener('resize', handleWindowResize)
    }, [])

    const handleUseCaseButtonClick = (useCaseIndex: number): void => {
        setActiveExampleIndex(useCaseIndex)
        telemetryService.log('CloudCodeInsightsGetStartedUseCase')
    }

    const isMobileLayout = windowSize <= 900

    return (
        <Card as={CardBody} className={classNames(styles.root, { [styles.rootMobile]: isMobileLayout })}>
            <div className={styles.section}>
                <Typography.H2>How engineering teams and leaders use Code Insights</Typography.H2>

                <p className="text-muted">
                    We've created a few common simple insights to show you what the tool can do.{' '}
                    <Link to="/help/code_insights">Explore more use cases.</Link>
                </p>

                {!isMobileLayout && (
                    <ul className={styles.list}>
                        {EXAMPLES.map((example, index) => (
                            <li key={index}>
                                <Button
                                    outline={index !== activeExampleIndex}
                                    variant={index === activeExampleIndex ? 'secondary' : undefined}
                                    className={classNames(styles.button, {
                                        [styles.buttonActive]: index === activeExampleIndex,
                                    })}
                                    onClick={() => handleUseCaseButtonClick(index)}
                                >
                                    {example.description}
                                    <svg
                                        className={styles.buttonSvg}
                                        viewBox="0 0 16 36"
                                        fill="none"
                                        xmlns="http://www.w3.org/2000/svg"
                                    >
                                        <path d="M0.443444 0H0V36H0.443444C2.04896 36 3.55682 35.229 4.49684 33.9275L15.1543 19.171C15.6591 18.472 15.6591 17.528 15.1543 16.829L4.49684 2.07255C3.55682 0.770985 2.04896 0 0.443444 0Z" />
                                    </svg>
                                </Button>
                            </li>
                        ))}
                    </ul>
                )}
            </div>

            {!isMobileLayout && (
                <CodeInsightExample
                    {...EXAMPLES[activeExampleIndex]}
                    className={styles.section}
                    telemetryService={telemetryService}
                />
            )}

            {isMobileLayout && <CodeInsightsExamplesSlider telemetryService={telemetryService} />}
        </Card>
    )
}
