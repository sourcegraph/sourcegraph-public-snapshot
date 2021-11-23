// @ts-check

/** @type {import('eslint').Rule.RuleModule} */
const config = {
  meta: {
    docs: {
      description: 'Checks for forbidden className literals. Recommends Wildcard components instead',
    },
  },

  create: function (context) {
    const configuration = context.options[0] || {}
    const forbiddenClassNames = new Map(configuration.forbid.map(({ className, message }) => [className, message]))

    return {
      JSXAttribute: node => {
        if (node.name.name !== 'className') {
          return
        }

        const classNames = node.value.value

        if (!classNames) {
          return
        }

        classNames.split(' ').forEach(className => {
          const message = forbiddenClassNames.get(className)
          if (message) {
            context.report({
              node,
              message: `"${className}" is forbidden. ${message}`,
            })
          }
        })
      },
    }
  },
}

module.exports = config
