import { useEffect } from 'react'

const mathjaxURL = 'https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js'
export const mathjaxElementId = 'MathJax-script'

/**
 * useMathJax enables rendering mathematical expressions on the page.
 *
 * @details
 * On component mount, useMathJax injects a script to load MathJax
 * from a dedicated CDN. This is the approach recommended in the
 * {@link https://docs.mathjax.org/en/latest/web/configuration.html#local-configuration-file | official documentation}.
 *
 * ```
 *  <script type="text/javascript" id="MathJax-script" async
 *      src="https://cdn.jsdelivr.net/npm/mathjax@3/es5/tex-mml-chtml.js">
 *  </script>
 * ````
 *
 * On component unmount, removes the script via useEffect destructor.
 */
export const useMathJax = () => {
    useEffect(() => {
        const mj = document.createElement('script')

        mj.setAttribute('type', 'text/javascript')
        // mj.setAttribute('src', mathjaxURL)
        mj.setAttribute('async', '')
        mj.setAttribute('id', mathjaxElementId)

        document.head.appendChild(mj)

        return () => {
            mj.remove()
        }
    }, [])

    // When working with dynamic HTML, it can happen that MathJax typesets the page
    // before the dynamic part containing mathematics is loaded, in which case we
    // need to trigger typesetting by ourselves.
    useEffect(() => {
        if (window.MathJax) {
            window.MathJax.typeset()
        }
    }, [window.MathJax])
}

declare global {
    interface Window {
        MathJax: { typeset: () => void } | undefined
    }
}
