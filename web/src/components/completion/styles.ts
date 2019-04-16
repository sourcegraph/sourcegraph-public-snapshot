import { CompletionWidgetClassProps } from '../../../../shared/src/components/completion/CompletionWidget'

const listItemClassName = 'completion-widget-dropdown__item d-flex align-items-center p-2'

export const COMPLETION_WIDGET_CLASS_PROPS: CompletionWidgetClassProps = {
    listClassName: 'completion-widget-dropdown d-block list-unstyled rounded p-0 m-0 mt-3',
    listItemClassName,
    selectedListItemClassName: 'completion-widget-dropdown__item--selected bg-primary',
    loadingClassName: listItemClassName,
    noResultsClassName: listItemClassName,
}
