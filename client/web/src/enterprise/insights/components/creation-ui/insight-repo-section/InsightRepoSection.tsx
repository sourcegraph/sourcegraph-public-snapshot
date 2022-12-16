import { ChangeEvent, FC, PropsWithChildren, ReactElement } from 'react'

import classNames from 'classnames'
import LinkExternalIcon from 'mdi-react/OpenInNewIcon'

import { EditorHint, QueryChangeSource, QueryState, SearchPatternType } from '@sourcegraph/search'
import { SyntaxHighlightedSearchQuery } from '@sourcegraph/search-ui'
import { Button, Code, Input, Label, InputElement, InputErrorMessage, InputDescription } from '@sourcegraph/wildcard'

import { useExperimentalFeatures } from '../../../../../stores'
import { CreateInsightFormFields } from '../../../pages/insights/creation/search-insight'
import {
    FormGroup,
    getDefaultInputProps,
    getDefaultInputStatus,
    getDefaultInputError,
    RepositoriesField,
    useFieldAPI,
    MonacoField,
} from '../../form'
import { MonacoPreviewLink } from '../../form/monaco-field'

import styles from './InsightRepoSection.module.scss'

interface RepoSettingSectionProps {
    repositories: useFieldAPI<CreateInsightFormFields['repositories']>
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
    repoMode: useFieldAPI<CreateInsightFormFields['repoMode']>
}

/**
 * Main entry point for the repositories insight setting section for the creation UI.
 * It contains all possible variation for the repo setting section based on feature and
 * experimental flags.
 */
export const RepoSettingSection: FC<RepoSettingSectionProps> = props => {
    const { repositories, repoQuery, repoMode } = props

    const repoUIVariation = useExperimentalFeatures(features => features.codeInsightsRepoUI)

    if (repoUIVariation === 'single-search-query') {
        return <SmartRepoSettingSection repoQuery={repoQuery} />
    }

    return <SearchQueryOrRepoListSection repoMode={repoMode} repoQuery={repoQuery} repositories={repositories} />
}

interface SmartRepoSettingSectionProps {
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
}

/**
 * Single smart search repo query field, this section is one of possible variation for the
 * strat-scoped insight repo query UI.
 */
export const SmartRepoSettingSection: FC<SmartRepoSettingSectionProps> = props => {
    const { repoQuery } = props

    return (
        <FormGroup name="insight repositories" title="Targeted repositories">
            <SmartSearchQueryRepoField repoQuery={repoQuery} />
        </FormGroup>
    )
}

interface SearchQueryOrRepoListSectionProps {
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
    repoMode: useFieldAPI<CreateInsightFormFields['repoMode']>
    repositories: useFieldAPI<CreateInsightFormFields['repositories']>
}

export const SearchQueryOrRepoListSection: FC<SearchQueryOrRepoListSectionProps> = props => {
    const { repoQuery, repoMode, repositories } = props

    return (
        <FormGroup
            name="insight repositories"
            title="Targeted repositories"
            contentClassName={styles.radioGroupSection}
        >
            <RadioGroupSection
                label="Repositories query"
                name="repoMode"
                labelId="smart-repo-search-query"
                value="search-query"
                checked={repoMode.input.value === 'search-query'}
                onChange={repoMode.input.onChange}
            >
                <SmartSearchQueryRepoField repoQuery={repoQuery} aria-labelledby="smart-repo-search-query" />
            </RadioGroupSection>

            <RadioGroupSection
                label="Explicit list of repositories"
                name="repoMode"
                labelId="strict-list-repo"
                value="urls-list"
                checked={repoMode.input.value === 'urls-list'}
                onChange={repoMode.input.onChange}
            >
                <Input
                    as={RepositoriesField}
                    message="Use a full repo URL (github.com/...). Separate repositories with comas"
                    placeholder="Example: github.com/sourcegraph/sourcegraph"
                    aria-labelledby="strict-list-repo"
                    {...getDefaultInputProps(repositories)}
                />
            </RadioGroupSection>
        </FormGroup>
    )
}

interface RadioGroupSectionProps {
    name: string
    label: string
    value: string
    checked: boolean
    labelId: string
    className?: string
    contentClassName?: string
    onChange: (event: ChangeEvent<HTMLInputElement>) => void
}

function RadioGroupSection(props: PropsWithChildren<RadioGroupSectionProps>): ReactElement {
    const { name, label, value, checked, labelId, children, onChange } = props

    return (
        <div className={styles.radioGroup}>
            {/* eslint-disable-next-line react/forbid-elements */}
            <input
                id={labelId}
                name={name}
                type="radio"
                value={value}
                checked={checked}
                className={styles.radioGroupInput}
                onChange={onChange}
            />
            <Label htmlFor={labelId} className={styles.radioGroupLabel}>
                {label}
            </Label>
            <div
                className={classNames(styles.radioGroupContent, {
                    [styles.radioGroupContentNonActive]: !checked,
                })}
            >
                {children}
            </div>
        </div>
    )
}

interface SmartSearchQueryRepoFieldProps {
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
    label?: string
    disabled?: boolean
    'aria-labelledby'?: string
}

function SmartSearchQueryRepoField(props: SmartSearchQueryRepoFieldProps): ReactElement {
    const { repoQuery, label, 'aria-labelledby': ariaLabelledby } = props

    const { value, onChange, disabled, ...attributes } = repoQuery.input

    const handleChipSuggestions = (chip: SmartRepoQueryChip): void => {
        const nextQueryValue = `${value.query} ${chip.query}`.trimStart()
        onChange({ query: nextQueryValue, hint: EditorHint.Focus })
    }

    const handleOnChange = (queryState: QueryState): void => {
        if (queryState.query !== value.query) {
            onChange({ query: queryState.query, changeSource: QueryChangeSource.userInput })
        }
    }

    const LabelComponent = label ? Label : 'div'

    return (
        <div>
            <LabelComponent className={styles.repoLabel} id="search-repo-query">
                {label && <span className={styles.repoLabelText}>Repositories query</span>}

                <InputElement
                    as={MonacoField}
                    queryState={value}
                    status={getDefaultInputStatus(repoQuery, value => value.query)}
                    placeholder="Example: repo:^github\.com/sourcegraph/sourcegraph$"
                    aria-labelledby={ariaLabelledby ?? 'search-repo-query'}
                    className={styles.repoInput}
                    onChange={handleOnChange}
                    disabled={disabled}
                    {...attributes}
                />

                <MonacoPreviewLink
                    query={value.query}
                    patternType={SearchPatternType.standard}
                    className={styles.repoLabelPreviewLink}
                    tabIndex={disabled ? -1 : 0}
                >
                    <LinkExternalIcon size={18} />
                </MonacoPreviewLink>
            </LabelComponent>

            <SmartRepoQueryChips disabled={disabled} onChipClick={handleChipSuggestions} />

            {getDefaultInputError(repoQuery) && (
                <InputErrorMessage message={getDefaultInputError(repoQuery)} className="mt-2 mb-2" />
            )}

            <InputDescription>
                <ul>
                    <li>
                        Hint: you can use regular expressions within each of the <Code weight="bold">before:</Code>{' '}
                        available filters
                    </li>
                    <li>
                        Datapoints will be automatically backfilled using the list <Code weight="bold">before:</Code> of
                        repositories resulting from today’s search. Future data points will use the list refreshed for
                        every snapshot.
                    </li>
                </ul>
            </InputDescription>
        </div>
    )
}

interface SmartRepoQueryChip {
    id: string
    query: string
}

const CHIP_QUERIES: SmartRepoQueryChip[] = [
    { id: '1', query: 'repo:' },
    { id: '2', query: '-repo:' },
    { id: '3', query: 'AND' },
    { id: '4', query: 'OR' },
    { id: '5', query: 'NOT' },
    { id: '6', query: 'select:repo' },
    { id: '7', query: 'repo:has.path()' },
    { id: '8', query: 'repo:has.file()' },
    { id: '9', query: 'repo:has.commit.after()' },
    { id: '10', query: 'repo:.*' },
]

interface SmartRepoQueryChipsProps {
    disabled?: boolean
    onChipClick: (chip: SmartRepoQueryChip) => void
}

function SmartRepoQueryChips(props: SmartRepoQueryChipsProps): ReactElement {
    const { disabled, onChipClick } = props

    return (
        <ul className={styles.chipsList}>
            {CHIP_QUERIES.map(chip => (
                <li key={chip.id}>
                    <Button
                        type="button"
                        tabIndex={disabled ? -1 : 0}
                        disabled={disabled}
                        className={styles.queryChip}
                        onClick={() => onChipClick(chip)}
                    >
                        <SyntaxHighlightedSearchQuery
                            query={chip.query}
                            searchPatternType={SearchPatternType.standard}
                        />
                    </Button>
                </li>
            ))}
        </ul>
    )
}
