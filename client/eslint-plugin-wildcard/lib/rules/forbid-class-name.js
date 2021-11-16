// @ts-check

/** @type {import('eslint').Rule.RuleModule} */
const config = {
  meta: {
    docs: {
      description: 'Checks for forbidden className literals. Recommends Wildcard components instead',
      recommended: true,
    },
  },

  create: function (context) {
    const configuration = context.options[0] || {}
    const classNameReplacements = new Map(
      configuration.forbid.map(({ className, component }) => [className, component])
    )

    return {
      JSXOpeningElement: node => {
        const classNames = node.attributes.find(attr => attr?.name?.name === 'className')?.value?.value || ''

        if (!classNames) {
          return
        }

        classNames.split(' ').forEach(className => {
          const replacement = classNameReplacements.get(className)
          if (replacement) {
            context.report({
              node,
              message: `Do not use the "${className}" class. Use the ${replacement} Wildcard component.`,
            })
          }
        })
      },
    }
  },
}

module.exports = config
