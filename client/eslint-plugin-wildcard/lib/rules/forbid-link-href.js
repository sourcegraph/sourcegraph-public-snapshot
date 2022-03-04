// @ts-check

/** @type {import('eslint').Rule.RuleModule} */
const config = {
  meta: {
    docs: {
      description: 'Checks for forbidden href values. Recommends alternative instead',
    },
    fixable: 'code',
  },

  create: function (context) {
    const configuration = context.options[0] || {}
    const forbiddenHrefs = configuration.forbid

    return {
      JSXAttribute: node => {
        if (node.name.name !== 'href' && node.name.name !== 'to') {
          return
        }

        const href = node.value.value

        if (!href) {
          return
        }

        const match = forbiddenHrefs.find(forbidden => href.startsWith(forbidden.href))

        if (match) {
          context.report({
            node,
            message: `"${href}" is forbidden. ${match.message}`,
            fix: fixer => fixer.replaceText(node.value, `"${href.replace(match.href, match.replaceWith)}"`),
          })
        }
      },
    }
  },
}

module.exports = config
