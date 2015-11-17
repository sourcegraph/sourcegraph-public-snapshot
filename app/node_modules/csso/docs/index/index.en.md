CSSO (CSS Optimizer) is a CSS minimizer unlike others. In addition to usual minification techniques it can perform structural optimization of CSS files, resulting in smaller file size compared to other minifiers.

## Minification (in a nutshell)

Safe transformations:

* Removal of whitespace
* Removal of trailing `;`
* Removal of comments
* Removal of invalid `@charset` and `@import` declarations
* Minification of color properties
* Minification of `0`
* Minification of multi-line strings
* Minification of the `font-weight` property

Structural optimizations:

* Merging blocks with identical selectors
* Merging blocks with identical properties
* Removal of overridden properties
* Removal of overridden shorthand properties
* Removal of repeating selectors
* Partial merging of blocks
* Partial splitting of blocks
* Removal of empty ruleset and at-rule
* Minification of `margin` and `padding` properties

The minification techniques are described in detail in the [detailed description](../description/description.en.md).

## Authors

* initial idea&nbsp;— Vitaly Harisov (<vitaly@harisov.name>)
* implementation&nbsp;— Sergey Kryzhanovsky (<skryzhanovsky@ya.ru>)
* english translation&nbsp;— Leonid Khachaturov (<leonidkhachaturov@gmail.com>)
* japanese translation&nbsp;— Koji Ishimoto (<ijok.ijok@gmail.com>)
* korean translation&nbsp;— Wankyu Kim (<wankyu19@gmail.com>)

## Feedback

Please report issues on [Github](https://github.com/css/csso/issues).

For feedback, suggestions, etc. write to <skryzhanovsky@ya.ru>.

## License

* CSSO is licensed under [MIT](https://github.com/css/csso/blob/master/MIT-LICENSE.txt)
