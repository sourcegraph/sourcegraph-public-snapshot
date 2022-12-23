import { SEMANTIC_COLORS } from './constants'

export const preventDefault = <E extends React.SyntheticEvent>(event: E): E => {
    event.preventDefault()
    return event
}

const isStyleRule = (rule: CSSRule): rule is CSSStyleRule => rule.type === 1

// https://css-tricks.com/how-to-get-all-custom-properties-on-a-page-in-javascript/
const getCSSCustomProperties = (): string[] => [
    ...new Set(
        [...document.styleSheets].flatMap(sheet =>
            [...sheet.cssRules].filter(isStyleRule).flatMap(rule => {
                const variables = [...rule.style]
                    .map(propertyName => propertyName.trim())
                    .filter(propertyName => propertyName.startsWith('--'))
                return variables
            })
        )
    ),
]

type SemanticColorBase = typeof SEMANTIC_COLORS[number]
type SemanticColorVariant = `${SemanticColorBase}-${number}`
type SemanticColor = SemanticColorBase | SemanticColorVariant

const isSemanticColor =
    (colorPattern: RegExp) =>
    (value: string): value is SemanticColor =>
        colorPattern.test(value)

export const getSemanticColorVariables = (): SemanticColor[] => {
    const properties = getCSSCustomProperties()
    return SEMANTIC_COLORS.flatMap(color => {
        const colorMatcher = isSemanticColor(new RegExp(`^--${color}(-\\d)?$`))
        return properties.filter(colorMatcher).sort()
    })
}
