module.exports = grammar({
  name: 'test_grammar',

  extras: $ => [/\s/, $.comment],

  rules: {
    expression: $ => choice(
      $.sum,
      $.number,
      $.variable,
      seq('(', $.expression, ')')
    ),
    sum: $ => prec.left(1, seq(field('left', $.expression), '+', field('right', $.expression))),
    number: $ => /\d+/,
    comment: $ => token(seq('//', /.*/)),
    variable: $ => /[a-zA-Z]\\w*/,
  }
});
