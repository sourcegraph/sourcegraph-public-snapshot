declare module 'uglifyjs-webpack-plugin' {
    import * as webpack from 'webpack'
    const UglifyJsPlugin: typeof webpack.optimize.UglifyJsPlugin
    export default UglifyJsPlugin
}
