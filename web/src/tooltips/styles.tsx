const boxShadow = '0 3px 12px rgba(27,31,35,0.15)'
const borderRadius = '3px'
const normalFontFamily = `-apple-system, BlinkMacSystemFont, "Segoe UI", Helvetica, Arial, sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol"`
const codeFontFamily = `"SFMono-Regular", Consolas, "Liberation Mono", Menlo, Courier, monospace`
import * as colors from 'sourcegraph/util/colors'

export const tooltip = {
    backgroundColor: colors.backgroundColor,
    maxWidth: '500px',
    minWidth: '400px',
    border: `solid 1px ${colors.borderColor}`,
    fontFamily: normalFontFamily,
    color: colors.normalFontColor,
    fontSize: '12px',
    zIndex: 100,
    position: 'absolute',
    overflow: 'auto',
    borderRadius,
    boxShadow
}

export const tooltipTitle = {
    fontFamily: codeFontFamily,
    wordWrap: 'break-word',
    whiteSpace: 'pre-wrap',
    marginLeft: '0px',
    marginRight: '32px',
    padding: '0px'
}

export const divider = {
    borderBottom: `1px solid ${colors.borderColor}`,
    paddingLeft: '8px',
    paddingRight: '8px',
    paddingBottom: '16px',
    paddingTop: '16px',
    lineHeight: '16px'
}

export const tooltipDoc = {
    paddingTop: '16px',
    paddingLeft: '8px',
    paddingRight: '16px',
    paddingBottom: '16px',
    maxHeight: '150px',
    overflow: 'auto',
    marginBottom: '0px',
    whiteSpace: 'pre-wrap',
    fontFamily: normalFontFamily,
    borderBottom: `1px solid ${colors.borderColor}`
}

export const tooltipActions = {
    display: 'flex',
    textAlign: 'center'
}

export const tooltipAction = {
    flex: 1,
    cursor: 'pointer',
    textDecoration: 'none',
    paddingTop: '8px',
    paddingBottom: '8px',
    paddingLeft: '8px',
    paddingRight: '8px',
    color: colors.actionFontColor
}

export const tooltipActionNotLast = {
    borderRight: `1px solid ${colors.borderColor}`
}

export const tooltipMoreActions = {
    fontStyle: 'italic',
    color: colors.actionFontColor,
    paddingTop: '8px',
    paddingBottom: '8px',
    paddingLeft: '16px',
    paddingRight: '16px'
}

export const fileNavButton = {
    borderTopLeftRadius: 0,
    borderBottomLeftRadius: 0,
    color: 'black',
    textDecoration: 'none'
}

export const sourcegraphIcon = {
    width: '16px',
    position: 'absolute',
    left: '16px',
    marginBottom: '2px',
    verticalAlign: 'middle'
}

export const closeIcon = {
    width: '12px',
    position: 'absolute',
    top: '18px',
    right: '16px',
    verticalAlign: 'middle',
    cursor: 'pointer'
}

export const loadingTooltip = {
    padding: '16px'
}

export const definitionIcon = {
    marginRight: '12px'
}

export const referencesIcon = {
    marginRight: '12px'
}

export const searchIcon = {
    marginRight: '12px',
    position: 'relative',
    top: '2px'
}
