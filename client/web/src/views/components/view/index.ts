import { ViewContent, ViewErrorContent, ViewLoadingContent } from './content/ViewContent'
import { View } from './View'

const Root = View
const Content = ViewContent
const LoadingContent = ViewLoadingContent
const ErrorContent = ViewErrorContent

export {
    View,
    ViewContent,
    ViewLoadingContent,
    ViewErrorContent,
    // For * as View import style
    Root,
    Content,
    LoadingContent,
    ErrorContent,
}
