import {
    type ChangeEvent,
    type FC,
    type MutableRefObject,
    type PropsWithChildren,
    type ReactElement,
    type ReactNode,
    useRef,
} from 'react'

import { gql, useQuery } from '@apollo/client'
import classNames from 'classnames'
import LinkExternalIcon from 'mdi-react/OpenInNewIcon'

import { SyntaxHighlightedSearchQuery } from '@sourcegraph/branded'
import { SearchPatternType } from '@sourcegraph/shared/src/graphql-operations'
import { EditorHint, QueryChangeSource, type QueryState } from '@sourcegraph/shared/src/search'
import { useExperimentalFeatures } from '@sourcegraph/shared/src/settings/settings'
import {
    Button,
    Code,
    Label,
    InputElement,
    InputErrorMessage,
    InputDescription,
    InputStatus,
    useDebounce,
    Link,
    FormGroup,
    type useFieldAPI,
    getDefaultInputProps,
    getDefaultInputStatus,
    getDefaultInputError,
} from '@sourcegraph/wildcard'

import type {
    InsightRepositoriesCountResult,
    InsightRepositoriesCountVariables,
} from '../../../../../graphql-operations'
import type { CreateInsightFormFields } from '../../../pages/insights/creation/search-insight'
import { getRepoQueryPreview, RepositoriesField, Field } from '../../form'
import { PreviewLink } from '../../form/field'

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
            <SmartSearchQueryRepoField repoQuery={repoQuery} label="Repositories query" />
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
                labelContent={<RepositoriesCount repoQuery={repoQuery} />}
                labelClassName="d-flex justify-content-between"
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
                <RepositoriesURLsPicker repositories={repositories} aria-labelledby="strict-list-repo" />
            </RadioGroupSection>
        </FormGroup>
    )
}

interface RepositoriesURLsPickerProps {
    repositories: useFieldAPI<CreateInsightFormFields['repositories']>
    'aria-labelledby': string
}

function RepositoriesURLsPicker(props: RepositoriesURLsPickerProps): ReactElement {
    const { repositories, 'aria-labelledby': ariaLabelledby } = props

    const { value, disabled, ...attributes } = getDefaultInputProps(repositories)
    const fieldValue = disabled ? [] : value

    return (
        <RepositoriesField
            id="repositories-id"
            description="Find and choose at least 1 repository to run insight"
            placeholder="Search repositories..."
            aria-labelledby={ariaLabelledby}
            aria-invalid={!!repositories.meta.error}
            value={fieldValue}
            {...attributes}
        />
    )
}

interface RadioGroupSectionProps {
    name: string
    label: string
    labelContent?: ReactNode
    value: string
    checked: boolean
    labelId: string
    className?: string
    labelClassName?: string
    contentClassName?: string
    onChange: (event: ChangeEvent<HTMLInputElement>) => void
}

function RadioGroupSection(props: PropsWithChildren<RadioGroupSectionProps>): ReactElement {
    const { name, label, value, checked, labelId, labelContent, labelClassName, children, onChange } = props

    return (
        <div className={styles.radioGroup}>
            {/*
                Standard wildcard input doesn't provide a simple layout for the radio element,
                in order to have custom layout in the repo control we have to use native input
                with custom styles around spacing and layout
            */}
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
            <Label htmlFor={labelId} className={classNames(labelClassName, styles.radioGroupLabel)}>
                {label}

                {labelContent && (
                    <span className={classNames('ml-auto', { [styles.radioGroupContentNonActive]: !checked })}>
                        {labelContent}
                    </span>
                )}
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

const EMPTY_QUERY_STATE: QueryState = { query: '' }

interface SmartSearchQueryRepoFieldProps {
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
    label?: string
    disabled?: boolean
    'aria-labelledby'?: string
}

function SmartSearchQueryRepoField(props: SmartSearchQueryRepoFieldProps): ReactElement {
    const { repoQuery, label, 'aria-labelledby': ariaLabelledby } = props

    const { value, onChange, disabled, ...attributes } = repoQuery.input

    // We have to have mutable value here for the disabled state in order to prevent
    // any updates from the repo query field when this field is disabled, see handleOnChange
    // callback. Mutable value is needed because codemirror doesn't update callback in a sync
    // way and even if we have disabled: true code mirror still preserves the prev callback
    // where we still have disabled: false in the scope.
    const disabledRefValue = useMutableRefValue(disabled)

    const handleChipSuggestions = (chip: SmartRepoQueryChip): void => {
        const nextQueryValue = `${value.query} ${chip.query}`.trimStart()
        onChange({ query: nextQueryValue, hint: EditorHint.Focus })
    }

    const handleOnChange = (queryState: QueryState): void => {
        if (queryState.query !== value.query && !disabledRefValue.current) {
            onChange({ query: queryState.query, changeSource: QueryChangeSource.userInput })
        }
    }

    const queryState = disabled ? EMPTY_QUERY_STATE : value
    const previewQuery = value.query ? getRepoQueryPreview(value.query) : value.query
    const fieldStatus = getDefaultInputStatus(repoQuery, value => value.query)
    const LabelComponent = label ? Label : 'div'

    return (
        <div>
            <LabelComponent className={styles.repoLabel} id="search-repo-query">
                {label && (
                    <span className={styles.repoLabelText}>
                        Repositories query
                        <RepositoriesCount repoQuery={repoQuery} />
                    </span>
                )}

                <InputElement
                    as={Field}
                    queryState={queryState}
                    status={fieldStatus}
                    placeholder="Example: repo:sourcegraph/*"
                    aria-labelledby={ariaLabelledby ?? 'search-repo-query'}
                    className={styles.repoInput}
                    onChange={handleOnChange}
                    disabled={disabled}
                    aria-busy={fieldStatus === InputStatus.loading}
                    aria-invalid={!!repoQuery.meta.error}
                    {...attributes}
                />

                <PreviewLink
                    query={previewQuery}
                    patternType={SearchPatternType.standard}
                    className={styles.repoLabelPreviewLink}
                    tabIndex={disabled ? -1 : 0}
                >
                    <LinkExternalIcon size={18} />
                </PreviewLink>
            </LabelComponent>

            <SmartRepoQueryChips disabled={disabled} onChipClick={handleChipSuggestions} />

            {getDefaultInputError(repoQuery) && (
                <InputErrorMessage message={getDefaultInputError(repoQuery)} className="mt-2 mb-2" />
            )}

            <InputDescription>
                <ul>
                    <li>
                        Hint: you can use regular expressions within each of the <Code weight="bold">repo:</Code>{' '}
                        <Link
                            to="/help/code_search/reference/queries#repository-search"
                            target="_blank"
                            rel="noopener noreferrer"
                        >
                            filters
                        </Link>
                    </li>
                    <li>
                        Data points will be automatically backfilled using the list of repositories resulting from
                        today’s search. Future data points will use the list refreshed for every snapshot.
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

function useMutableRefValue<T>(value: T): MutableRefObject<T> {
    const valueRef = useRef<T>(value)
    valueRef.current = value

    return valueRef
}

const REPOSITORIES_COUNT_GQL = gql`
    query InsightRepositoriesCount($query: String!) {
        previewRepositoriesFromQuery(query: $query) {
            numberOfRepositories
        }
    }
`

interface RepositoriesCountProps {
    repoQuery: useFieldAPI<CreateInsightFormFields['repoQuery']>
    className?: string
}

function RepositoriesCount(props: RepositoriesCountProps): ReactElement {
    const { repoQuery, className } = props

    const query = useDebounce(!repoQuery.input.disabled ? repoQuery.input.value.query : '', 500)

    const { data } = useQuery<InsightRepositoriesCountResult, InsightRepositoriesCountVariables>(
        REPOSITORIES_COUNT_GQL,
        { skip: repoQuery.input.disabled, variables: { query } }
    )

    const repositoriesNumber = !repoQuery.input.disabled
        ? data?.previewRepositoriesFromQuery.numberOfRepositories ?? 0
        : 0

    return (
        <span className={classNames(className, 'text-muted font-weight-normal')}>
            Repositories count: {repositoriesNumber}
        </span>
    )
}
