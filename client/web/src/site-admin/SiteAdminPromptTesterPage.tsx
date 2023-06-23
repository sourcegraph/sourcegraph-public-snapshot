import React, { useCallback, useEffect, useState } from 'react'

import { mdiPlus } from '@mdi/js'

import { useTemporarySetting } from '@sourcegraph/shared/src/settings/temporary/useTemporarySetting'
import { TelemetryProps } from '@sourcegraph/shared/src/telemetry/telemetryService'
import { PromptVersion } from '@sourcegraph/shared/src/util/prompt-tester'
import {
    Button,
    Container,
    ErrorAlert,
    H3,
    Icon,
    Input,
    Label,
    LoadingSpinner,
    PageHeader,
    Select,
    TextArea,
    Tooltip,
} from '@sourcegraph/wildcard'

import { CompletionRequest, getCodyCompletionOneShot } from '../cody/search/api'
import { PageTitle } from '../components/PageTitle'

import styles from './SiteAdminPromptTesterPage.module.scss'

export interface SiteAdminPromptTesterPageProps extends TelemetryProps {}

const DEFAULT_VARIATION_COUNT = 1

const DEFAULT_TEMPERATURE = 0.2
const DEFAULT_MAX_TOKENS_TO_SAMPLE = 1000

/**
 * If the prompt is in the format "Human: <text>\n\nAssistant: <text>\n\nHuman: <text>\n\nAssistant: <text>\n\n...",
 * then the returned array will have alternating human and assistant messages.
 * Otherwise, the returned array will have a single human message.
 */
function buildMessages(prompt: string): { speaker: 'human' | 'assistant'; text: string }[] {
    if (prompt.startsWith('\n\nHuman: ')) {
        return prompt.split('\n\n').map(line => {
            const [speaker, text] = line.split(': ')
            return { speaker: speaker.toLowerCase() as 'human' | 'assistant', text }
        })
    }
    return [
        { speaker: 'human', text: prompt },
        { speaker: 'assistant', text: '' },
    ]
}

async function getResults(
    promptVersions: PromptVersion[],
    variationCount: number,
    abortSignal: AbortSignal
): Promise<string[][]> {
    const resultPromises: Promise<string[]>[] = promptVersions.map(async promptVersion => {
        const params: CompletionRequest = {
            messages: buildMessages(promptVersion.prompt),
            temperature: promptVersion.temperature,
            maxTokensToSample: promptVersion.maxTokensToSample,
            topK: -1, // default value (source: https://console.anthropic.com/docs/api/reference)
            topP: -1, // default value
        }
        const completionPromises: Promise<string>[] = []
        for (let index = 0; index < variationCount; index++) {
            completionPromises.push(getCodyCompletionOneShot(params, abortSignal))
        }
        return Promise.all(completionPromises)
    })
    return Promise.all(resultPromises)
}

export const SiteAdminPromptTesterPage: React.FunctionComponent<
    React.PropsWithChildren<SiteAdminPromptTesterPageProps>
> = ({ telemetryService }) => {
    const [promptVersions, setPromptVersions] = useTemporarySetting('admin.promptTester.promptVersions', [
        {
            prompt: '',
            provider: 'anthropic',
            temperature: DEFAULT_TEMPERATURE,
            maxTokensToSample: DEFAULT_MAX_TOKENS_TO_SAMPLE,
        },
        {
            prompt: '',
            provider: 'openai',
            temperature: DEFAULT_TEMPERATURE,
            maxTokensToSample: DEFAULT_MAX_TOKENS_TO_SAMPLE,
        },
    ])
    const [results, setResults] = useState<string[][] | undefined>(undefined)
    const [error, setError] = useState<Error>()
    const [loadingResults, setLoadingResults] = useState<boolean>(false)
    const [variationCount, setVariationCount] = useTemporarySetting(
        'admin.promptTester.variationCount',
        DEFAULT_VARIATION_COUNT
    )
    const abortController = new AbortController()

    useEffect(() => {
        telemetryService.logPageView('SiteAdminPromptTester')
    }, [telemetryService])

    const runTests = useCallback(async () => {
        setLoadingResults(true)

        // Do stuff
        try {
            if (promptVersions && variationCount) {
                setResults(await getResults(promptVersions, variationCount, abortController.signal))
            }
        } catch (error) {
            setError(error)
        }

        setLoadingResults(false)
    }, [promptVersions, variationCount, abortController.signal])

    const addPromptVersion = useCallback(() => {
        setPromptVersions(promptVersions =>
            promptVersions
                ? [
                      ...promptVersions,
                      {
                          prompt: '',
                          provider: 'anthropic',
                          temperature: DEFAULT_TEMPERATURE,
                          maxTokensToSample: DEFAULT_MAX_TOKENS_TO_SAMPLE,
                      },
                  ]
                : undefined
        )
    }, [setPromptVersions])

    const setPromptVersion = useCallback(
        (index: number, newPV: PromptVersion) => {
            setPromptVersions(pvs =>
                pvs ? pvs.map((oldPV, currentIndex) => (currentIndex === index ? newPV : oldPV)) : undefined
            )
        },
        [setPromptVersions]
    )

    const removePromptVersion = useCallback(
        (index: number) => {
            setPromptVersions(pvs =>
                pvs ? pvs.filter((promptVersion, currentIndex) => currentIndex !== index) : undefined
            )
        },
        [setPromptVersions]
    )

    const onRunButtonClick = (): void => {
        runTests().catch(() => {})
    }

    return (
        <div className="site-admin-prompt-tester-page">
            <PageTitle title="Prompt tester - Admin" />
            <PageHeader
                path={[{ text: 'Prompt tester' }]}
                headingElement="h2"
                description={
                    <>
                        This is a tool to test prompts for Cody.
                        <br />
                        Add your prompts, select the LLMs, and click Run. Then compare the quality of the results.
                    </>
                }
                className="mb-3"
            />
            <Container className="mb-3">
                <H3>Prompts</H3>
                <div className={styles.promptVersion}>
                    <div>#</div>
                    <div>Prompt</div>
                    <div>Provider</div>
                    <div>Temperature</div>
                    <div>Output</div>
                </div>
                {promptVersions && promptVersions.length === 0 && <div>No prompts yet</div>}
                {!!promptVersions &&
                    promptVersions.map((promptVersion, index) =>
                        formatPromptVersion(promptVersion, index, setPromptVersion, removePromptVersion)
                    )}
                <Button variant="success" onClick={addPromptVersion}>
                    <Icon aria-hidden={true} svgPath={mdiPlus} /> Add prompt
                </Button>
            </Container>
            <div className={styles.variationCountContainer}>
                <Label htmlFor="variation-count">Number of variations to produce for each prompt:</Label>
                <Input
                    id="variation-count"
                    className={styles.variationCountInput}
                    type="number"
                    min={1}
                    max={10}
                    value={variationCount ?? 1}
                    onChange={event => setVariationCount(parseInt(event.target.value, 10))}
                />
            </div>
            <Button variant="primary" className={styles.runButton} onClick={onRunButtonClick} disabled={loadingResults}>
                Run tests
            </Button>

            <Container className="mb-3">
                <H3>Results</H3>
                {error && !loadingResults && <ErrorAlert error={error} />}
                {loadingResults && !error && <LoadingSpinner />}
                {results && (
                    <div>
                        {results.map((completions, resultIndex) => (
                            <div className={styles.resultRow} key={resultIndex}>
                                <div>{resultIndex + 1}</div>
                                {completions.map(completion => (
                                    <div key={completion}>{completion}</div>
                                ))}
                            </div>
                        ))}
                    </div>
                )}
            </Container>
        </div>
    )
}

function formatPromptVersion(
    promptVersion: PromptVersion,
    index: number,
    setPromptVersion: (index: number, promptVersion: PromptVersion) => void,
    removePromptVersion: (id: number) => void
): React.ReactNode {
    const onPromptChange = (event: React.ChangeEvent<HTMLTextAreaElement>): void =>
        setPromptVersion(index, {
            ...promptVersion,
            prompt: event.target.value,
        })
    const onProviderChange = (event: React.ChangeEvent<HTMLSelectElement>): void =>
        setPromptVersion(index, {
            ...promptVersion,
            provider: event.target.value as 'anthropic' | 'openai',
        })
    const onTemperatureChange = (temperature: number): void =>
        setPromptVersion(index, {
            ...promptVersion,
            temperature,
        })
    const onMaxTokensToSampleChange = (maxTokensToSample: number): void =>
        setPromptVersion(index, {
            ...promptVersion,
            maxTokensToSample,
        })
    return (
        <div className={styles.promptVersion}>
            <div>{index + 1}</div>
            <TextArea value={promptVersion.prompt} onChange={onPromptChange} placeholder="Paste prompt here" />
            <Select aria-labelledby="Provider" value={promptVersion.provider} onChange={onProviderChange}>
                <option value="anthropic">Anthropic</option>
                <option value="openai">OpenAI</option>
            </Select>
            <Tooltip content="Temperature. The higher this value, the more creative the results will be.">
                <Input
                    id={`temperature-${index}`}
                    type="number"
                    step={0.1}
                    min={0}
                    max={1}
                    value={promptVersion.temperature}
                    onChange={event => onTemperatureChange(parseFloat(event.target.value))}
                />
            </Tooltip>
            <Tooltip content="Max tokens to sample. The higher this value, the longer the results will be.">
                <Input
                    id={`max-tokens-${index}`}
                    type="number"
                    min={1}
                    max={5000}
                    value={promptVersion.maxTokensToSample}
                    onChange={event => onMaxTokensToSampleChange(parseInt(event.target.value, 10))}
                />
            </Tooltip>
            <Button variant="danger" onClick={() => removePromptVersion(index)}>
                Remove
            </Button>
        </div>
    )
}
