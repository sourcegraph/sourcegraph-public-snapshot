// @ts-check

const FORBIDDEN_URL_PREFIX = 'https://docs.sourcegraph.com'
const PREFERRED_URL_PREFIX = '/help'

/** @type {import('eslint').Rule.RuleModule} */
const config = {
  meta: {
    docs: {
      description:
        'Checks for forbidden docs.sourcegraph.com values. Recommends using the versioned "/help" redirect instead',
    },
    fixable: 'code',
  },

  create: function (context) {
    return {
      JSXAttribute: node => {
        if (node.name.name !== 'href' && node.name.name !== 'to') {
          return
        }

        const href = node.value.value

        if (!href || typeof href !== 'string') {
          return
        }

        if (href.startsWith(FORBIDDEN_URL_PREFIX)) {
          context.report({
            node,
            message: `"${href}" is forbidden. Use '/help' to redirect to the correct documentation for the current site version. See docs: https://sourcegraph.com/help/dev/how-to/documentation_implementation#linking-to-documentation-in-product`,
            fix: fixer =>
              fixer.replaceText(node.value, `"${href.replace(FORBIDDEN_URL_PREFIX, PREFERRED_URL_PREFIX)}"`),
          })
        }
      },
    }
  },
}

module.exports = config
