'use strict'

//------------------------------------------------------------------------------
// Rule Definition
//------------------------------------------------------------------------------

module.exports = {
  meta: {
    docs: {
      description: 'Check that /help links point to real, non-redirected pages',
      category: 'Best Practices',
      recommended: false,
    },
    schema: [
      {
        type: 'object',
        properties: {
          docsiteList: {
            type: 'array',
            items: {
              type: 'string',
            },
          },
        },
        additionalProperties: false,
      },
    ],
  },

  create: function (context) {
    // Build the set of valid pages. In order, we'll try to get this from:
    //
    // 1. The DOCSITE_LIST environment variable, which should be a newline
    //    separated list of pages, as outputted by `docsite ls`.
    // 2. The docsiteList rule option, which should be an array of pages.
    //
    // If neither of these are set, this rule will silently pass, so as not to
    // require docsite to be run when a user wants to run eslint in general.
    const pages = new Set()
    if (process.env.DOCSITE_LIST) {
      process.env.DOCSITE_LIST.split('\n').forEach(page => pages.add(page))
    } else if (context.options.length > 0) {
      context.options[0].docsiteList.forEach(page => pages.add(page))
    }

    // No pages were provided, so we'll return an empty object and do nothing.
    if (pages.size === 0) {
      return {}
    }

    // Return the object that will install the listeners we want. In this case,
    // we only need to look at JSX opening elements.
    //
    // Note that we could use AST selectors below, but the structure of the AST
    // makes that tricky: the identifer (Link or a) and attribute (to or href)
    // we use to identify an element of interest are siblings, so we'd probably
    // have to select on the identifier and have some ugly traversal code below
    // to check the attribute. It feels cleaner to do it this way with the
    // opening element as the context.
    return {
      JSXOpeningElement: node => {
        // Figure out what kind of element we have and therefore what attribute
        // we'd want to look for.
        let attrName
        if (node.name.name === 'Link') {
          attrName = 'to'
        } else if (node.name.name === 'a') {
          attrName = 'href'
        } else {
          // Anything that's not a link is uninteresting.
          return
        }

        // Go find the link target in the attribute array.
        const target = node.attributes.reduce(
          (target, attr) => target || (attr.name && attr.name.name === attrName ? attr.value.value : undefined),
          undefined
        )

        // Make sure the target points to a help link; if not, we don't need to
        // go any further.
        if (!target || !target.startsWith('/help/')) {
          return
        }

        // Strip off the /help/ prefix, any anchor, and any trailing slash, then
        // look up the resultant page in the pages set, bearing in mind that it
        // might point to a directory and we also need to look for any index
        // page that might exist.
        const destination = target.substring(6).split('#')[0].replace(/\/+$/, '')
        if (!pages.has(destination + '.md') && !pages.has(destination + '/index.md')) {
          context.report({
            node,
            message: 'Help link to non-existent page: {{ destination }}',
            data: { destination },
          })
        }
      },
    }
  },
}
