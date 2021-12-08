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

    /**
     * Extract string literals from JSX attributes
     * @param {any} attributeNode
     * @returns {[string]}
     */
    const extractStringLiteral = attributeNode => {
      if (attributeNode.type === 'JSXExpressionContainer') {
        return attributeNode.expression.arguments
          .filter(argument => argument.type === 'Literal')
          .flatMap(argument => argument.value.split(' '))
      }

      const literalValue = attributeNode.value

      if (!literalValue) {
        return
      }

      return literalValue.split(' ')
    }

    return {
      JSXAttribute: node => {
        if (node.name.name !== 'className') {
          return
        }

        const classNames = extractStringLiteral(node.value)

        if (!classNames) {
          return
        }

        classNames.forEach(className => {
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
