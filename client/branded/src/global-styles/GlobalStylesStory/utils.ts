import { SEMANTIC_COLORS } from './constants'

export const preventDefault = <E extends React.SyntheticEvent>(event: E): E => {
    event.preventDefault()
    return event
}

const isStyleRule = (rule: CSSRule): rule is CSSStyleRule => rule.type === 1

// https://css-tricks.com/how-to-get-all-custom-properties-on-a-page-in-javascript/
const getCSSCustomProperties = () => [
    ...new Set(
        [...document.styleSheets].reduce<string[]>(
            (finalArr, sheet) =>
                finalArr.concat(
                    [...sheet.cssRules].filter(isStyleRule).reduce<string[]>((totalVariables, rule) => {
                        const variables = [...rule.style]
                            .map(propName => propName.trim())
                            .filter(propName => propName.indexOf('--') === 0)
                        return [...totalVariables, ...variables]
                    }, [])
                ),
            []
        )
    ),
]

export const getSemanticColorVariables = () => {
    const properties = getCSSCustomProperties()
    return SEMANTIC_COLORS.flatMap(color =>
        properties.filter(customProperty => customProperty.match(`^--${color}(-\\d)?$`)).sort()
    )
}
