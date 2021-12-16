import { ViewBannerContent, ViewContent, ViewErrorContent, ViewLoadingContent } from './content/ViewContent'
import { View } from './View'

const Root = View
const Content = ViewContent
const LoadingContent = ViewLoadingContent
const ErrorContent = ViewErrorContent
const Banner = ViewBannerContent

export {
    View,
    ViewContent,
    ViewLoadingContent,
    ViewErrorContent,
    ViewBannerContent,
    // For * as View import style
    Root,
    Content,
    LoadingContent,
    ErrorContent,
    Banner,
}
