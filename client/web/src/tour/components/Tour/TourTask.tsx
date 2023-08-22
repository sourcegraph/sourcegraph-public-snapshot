import React, { useCallback, useContext, useEffect, useMemo, useState } from 'react'

import { mdiCheckCircle } from '@mdi/js'
import classNames from 'classnames'
import { CircularProgressbar } from 'react-circular-progressbar'
import { useNavigate } from 'react-router-dom'

import { ModalVideo } from '@sourcegraph/branded'
import { isExternalLink } from '@sourcegraph/common'
import { TourLanguage, type TourTaskStepType, type TourTaskType } from '@sourcegraph/shared/src/settings/temporary'
import { Button, Icon, Link, Text } from '@sourcegraph/wildcard'

import { ItemPicker } from '../ItemPicker'

import { TourContext } from './context'
import { TourNewTabLink } from './TourNewTabLink'
import { getTourTaskStepActionValue, isLanguageRequired } from './utils'
import { fromEvent, Observable, of, Subscriber, Subscription } from 'rxjs'

import styles from './Tour.module.scss'

import {
    AggregateStreamingSearchResults,
    LATEST_VERSION,
    messageHandlers,
    MessageHandlers, observeMessages,
    search,
    SearchEvent, SearchMatch, StreamSearchOptions, switchAggregateSearchResults,
} from '@sourcegraph/shared/src/search/stream'
import { SearchPatternType } from "@sourcegraph/shared/src/graphql-operations";

import { noop } from "lodash";
import { map } from "rxjs/operators";

type TourTaskProps = TourTaskType & {
    variant?: 'small'
}

/**
 * Tour task smart component. Handles all TourTaskStepType.type options.
 */
export const TourTask: React.FunctionComponent<React.PropsWithChildren<TourTaskProps>> = ({
    title,
    steps,
    completed,
    icon,
    variant,
    dataAttributes = {},
}) => {
    const [selectedStep, setSelectedStep] = useState<TourTaskStepType>()
    const [showLanguagePicker, setShowLanguagePicker] = useState(false)
    const { language, onLanguageSelect, onStepClick, onRestart } = useContext(TourContext)

    const handleLinkClick = useCallback(
        (
            event: React.MouseEvent<HTMLElement, MouseEvent> | React.KeyboardEvent<HTMLElement>,
            step: TourTaskStepType
        ) => {
            onStepClick(step, language)
            if (isLanguageRequired(step) && !language) {
                event.preventDefault()
                setShowLanguagePicker(true)
                setSelectedStep(step)
            }
        },
        [language, onStepClick]
    )

    const handleVideoToggle = useCallback(
        (isOpen: boolean, step: TourTaskStepType) => {
            if (!isOpen) {
                onStepClick(step, language)
            }
        },
        [language, onStepClick]
    )

    const onLanguageClose = useCallback(() => setShowLanguagePicker(false), [])

    const navigate = useNavigate()
    const handleLanguageSelect = useCallback(
        (language: TourLanguage) => {
            onLanguageSelect(language)
            setShowLanguagePicker(false)
            if (!selectedStep) {
                return
            }
            onStepClick(selectedStep, language)
            const url = getTourTaskStepActionValue(selectedStep, language)
            if (isExternalLink(url)) {
                window.open(url, '_blank')
            } else {
                navigate(url)
            }
        },
        [onStepClick, onLanguageSelect, selectedStep, navigate]
    )
    const attributes = useMemo(
        () =>
            Object.entries(dataAttributes).reduce(
                (result, [key, value]) => ({ ...result, [`data-${key}`]: value }),
                {}
            ),
        [dataAttributes]
    )

    if (showLanguagePicker) {
        return (
            <ItemPicker
                title="Please select a language:"
                className={classNames(variant !== 'small' && 'pl-2')}
                items={Object.values(TourLanguage)}
                onClose={onLanguageClose}
                onSelect={handleLanguageSelect}
            />
        )
    }

    const isMultiStep = steps.length > 1
    return (
        <div
            className={classNames(
                icon && [styles.task, variant === 'small' && styles.isSmall],
                !title && styles.noTitleTask
            )}
            {...attributes}
        >
            {icon && variant !== 'small' && <span className={styles.taskIcon}>{icon}</span>}
            <div className={classNames('flex-grow-1', variant !== 'small' && 'h-100 d-flex flex-column')}>
                {title && (
                    <div className="d-flex justify-content-between position-relative">
                        {icon && variant === 'small' && <span className={classNames(styles.taskIcon)}>{icon}</span>}
                        <Text className={styles.title}>{title}</Text>
                        {completed === 100 && (
                            <Icon size="sm" className="text-success" aria-label="Completed" svgPath={mdiCheckCircle} />
                        )}
                        {typeof completed === 'number' && completed < 100 && (
                            <CircularProgressbar
                                className={styles.progressBar}
                                strokeWidth={10}
                                value={completed || 0}
                            />
                        )}
                    </div>
                )}
                <ul
                    className={classNames(
                        styles.stepList,
                        'm-0',
                        variant !== 'small' && 'flex-grow-1 d-flex flex-column',
                        isMultiStep && styles.isMultiStep
                    )}
                >
                    {steps.map(step => (
                        <li key={step.id} className={classNames(styles.stepListItem, 'd-flex align-items-center')}>
                            {step.action.type === 'link' && (
                                <Link
                                    className="flex-grow-1"
                                    to={getTourTaskStepActionValue(step, language)}
                                    onClick={event => handleLinkClick(event, step)}
                                >
                                    {step.label}
                                </Link>
                            )}
                            {step.action.type === 'new-tab-link' && (
                                <TourNewTabLink
                                    step={step}
                                    variant={step.action.variant === 'button-primary' ? 'button' : 'link'}
                                    className={classNames('flex-grow-1')}
                                    to={getTourTaskStepActionValue(step, language)}
                                    onClick={handleLinkClick}
                                />
                            )}
                            {step.action.type === 'restart' && (
                                <div className="flex-grow">
                                    <Text className="m-0">{step.label}</Text>
                                    <div className="d-flex flex-column">
                                        <Button
                                            variant="link"
                                            className="align-self-start text-left pl-0 font-weight-normal"
                                            onClick={() => onRestart(step)}
                                        >
                                            {step.action.value}
                                        </Button>
                                    </div>
                                </div>
                            )}
                            {step.action.type === 'video' && (
                                <ModalVideo
                                    id={step.id}
                                    showCaption={true}
                                    title={step.label}
                                    className="flex-grow-1"
                                    titleClassName="shadow-none text-left p-0 m-0"
                                    src={getTourTaskStepActionValue(step, language)}
                                    onToggle={isOpen => handleVideoToggle(isOpen, step)}
                                />
                            )}
                            {step.action.type === 'SearchDynamicContentResults' && (
                                <SearchDynamicContentResultsTask snippets={step.action.snippets } />
                            )}
                            {(isMultiStep || !title) && step.isCompleted && (
                                <Icon
                                    size="md"
                                    className="text-success"
                                    aria-label="Completed step"
                                    svgPath={mdiCheckCircle}
                                />
                            )}
                        </li>
                    ))}
                </ul>
            </div>
        </div>
    )
}

interface SDCRTProps {
    step: TourTaskStepType
    snippets: string[]
}
const SearchDynamicContentResultsTask: React.FunctionComponent<React.PropsWithChildren<SDCRTProps>> = (props) => {
    const [selectedQuery, setSelectedQuery] = useState<string>('')

    const [repo, setRepo] = useState<string>('sourcegraph/sourcegraph')
    const [lang, setLang] = useState<string>('Go')

    useEffect(() => {

        let promises = []

        for (let snip of props.snippets) {
            promises.push(generateDynamicQueryPromise(snip, repo, lang))
        }

        Promise.any(promises).then(val => {
            setSelectedQuery(val)
        })
    }, [])

    return (
        <div>
            {selectedQuery && <Link
                // className={classNames('flex-grow-1')}
                to={buildSearchUri(buildBasicQuery(selectedQuery, repo, lang))}
                // onClick={handleLinkClick}
            >
                {buildBasicQuery(selectedQuery, repo, lang)}
            </Link>}

        </div>
    )
}

function buildSearchUri(query: String): String {
    return encodeURI(`/search?q=${query}`)
}

function buildBasicQuery(snippet: String, repo?: String, lang?: String): String {
    let query = `${snippet}`
    if (lang) {
        query = `lang:${lang} ${query}`
    }
    if (repo) {
        query = `repo:${repo} ${query}`
    }
    return query
}

function generateDynamicQueryPromise(snippet: String, repo?: String, lang?: String): Promise<String> {
    let query = `${buildBasicQuery(snippet, repo, lang)} timeout:5s count:1 select:content`
    console.log(`searching for ${query}`)

    return fetchStreamSuggestions(query).toPromise().then((res) => {
        if (res.length > 0) {
            return snippet
        }
        return Promise.reject()
    })
}


const noopHandler = <T extends SearchEvent>(
    type: T['type'],
    eventSource: EventSource,
    _observer: Subscriber<SearchEvent>
): Subscription => fromEvent(eventSource, type).subscribe(noop)

const firstMatchMessageHandlers: MessageHandlers = {
    ...messageHandlers,
    matches: (type, eventSource, observer) =>
        observeMessages(type, eventSource).subscribe(data => {
            observer.next(data)
            // Once we observer the first `matches` event, complete the stream and close the event source.
            observer.complete()
            eventSource.close()
        }),
    progress: noopHandler,
    filters: noopHandler,
    alert: noopHandler,
}

/** Initiates a streaming search, stop at the first `matches` event, and aggregate the results. */
function firstMatchStreamingSearch(
    queryObservable: Observable<string>,
    options: StreamSearchOptions
): Observable<AggregateStreamingSearchResults> {
    return search(queryObservable, options, firstMatchMessageHandlers).pipe(switchAggregateSearchResults)
}

function fetchStreamSuggestions(query: string, sourcegraphURL?: string): Observable<SearchMatch[]> {
    return firstMatchStreamingSearch(of(query), {
        version: LATEST_VERSION,
        patternType: SearchPatternType.standard,
        caseSensitive: false,
        trace: undefined,
        sourcegraphURL,
    }).pipe(map(suggestions => suggestions.results))
}
