# React CSS Modules

[![Travis build status](http://img.shields.io/travis/gajus/react-css-modules/master.svg?style=flat-square)](https://travis-ci.org/gajus/react-css-modules)
[![NPM version](http://img.shields.io/npm/v/react-css-modules.svg?style=flat-square)](https://www.npmjs.org/package/react-css-modules)
[![js-canonical-style](https://img.shields.io/badge/code%20style-canonical-blue.svg?style=flat-square)](https://github.com/gajus/canonical)

<img src='./.README/react-css-modules.png' height='150' />

React CSS Modules implement automatic mapping of CSS modules. Every CSS class is assigned a local-scoped identifier with a global unique name. CSS Modules enable a modular and reusable CSS!

- [CSS Modules](#css-modules)
    - [webpack `css-loader`](#webpack-css-loader)
- [What's the Problem?](#whats-the-problem)
- [The Implementation](#the-implementation)
- [Usage](#usage)
    - [Module Bundler](#module-bundler)
        - [webpack](#webpack)
            - [Development](#development)
            - [Production](#production)
        - [Browserify](#browserify)
    - [Extending Component Styles](#extending-component-styles)
    - [`styles` Property](#styles-property)
    - [Loops and Child Components](#loops-and-child-components)
    - [Decorator](#decorator)
    - [Options](#options)
        - [`allowMultiple`](#allowmultiple)
        - [`errorWhenNotFound`](#errorwhennotfound)
- [SASS, SCSS, LESS and other CSS Preprocessors](#sass-scss-less-and-other-css-preprocessors)
    - [Enable Sourcemaps](#enable-sourcemaps)
- [Class Composition](#class-composition)
    - [What Problems does Class Composition Solve?](#what-problems-does-class-composition-solve)
    - [Class Composition Using CSS Preprocessors](#class-composition-using-css-preprocessors)
- [SASS, SCSS, LESS and other CSS Preprocessors](#sass-scss-less-and-other-css-preprocessors)
- [Global CSS](#global-css)
- [Multiple CSS Modules](#multiple-css-modules)

## CSS Modules

[CSS Modules](https://github.com/css-modules/css-modules) are awesome. If you are not familiar with CSS Modules, it is a concept of using a module bundler such as [webpack](http://webpack.github.io/docs/) to load CSS scoped to a particular document. CSS module loader will generate a unique name for a each CSS class at the time of loading the CSS document ([Interoperable CSS](https://github.com/css-modules/icss) to be precise). To see CSS Modules in practice, [webpack-demo](https://css-modules.github.io/webpack-demo/).

In the context of React, CSS Modules look like this:

```js
import React from 'react';
import styles from './table.css';

export default class Table extends React.Component {
    render () {
        return <div className={styles.table}>
            <div className={styles.row}>
                <div className={styles.cell}>A0</div>
                <div className={styles.cell}>B0</div>
            </div>
        </div>;
    }
}
```

Rendering the component will produce a markup similar to:

```js
<div class="table__table___32osj">
    <div class="table__row___2w27N">
        <div class="table__cell___2w27N">A0</div>
        <div class="table__cell___1oVw5">B0</div>
    </div>
</div>
```

and a corresponding CSS file that matches those CSS classes.

Awesome!

### webpack `css-loader`

[CSS Modules](https://github.com/css-modules/css-modules) is a specification that can be implemented in multiple ways. `react-css-modules` leverages as existing CSS Modules implementation webpack [css-loader](https://github.com/webpack/css-loader#css-modules).

## What's the Problem?

webpack [css-loader](https://github.com/webpack/css-loader#css-modules) itself has several disadvantages:

* You have to use `camelCase` CSS class names.
* You have to use `styles` object whenever constructing a `className`.
* Mixing CSS Modules and global CSS classes is cumbersome.
* Reference to an undefined CSS Module resolves to `undefined` without a warning.

React CSS Modules component automates loading of CSS Modules using `styleName` property, e.g.

```js
import React from 'react';
import CSSModules from 'react-css-modules';
import styles from './table.css';

class Table extends React.Component {
    render () {
        return <div styleName='table'>
            <div styleName='row'>
                <div styleName='cell'>A0</div>
                <div styleName='cell'>B0</div>
            </div>
        </div>;
    }
}

export default CSSModules(Table, styles);
```

Using `react-css-modules`:

* You are not forced to use the `camelCase` naming convention.
* You do not need to refer to the `styles` object every time you use a CSS Module.
* There is clear distinction between global CSS and CSS Modules, e.g.

```js
<div className='global-css' styleName='local-module'></div>
```

* You are warned when `styleName` refers to an undefined CSS Module ([`errorWhenNotFound`](#errorwhennotfound) option).
* You can enforce use of a single CSS module per `ReactElement` ([`allowMultiple`](#allowmultiple) option).

## The Implementation

`react-css-modules` extends `render` method of the target component. It will use the value of `styleName` to look for CSS Modules in the associated styles object and will append the matching unique CSS class names to the `ReactElement` `className` property value.

[Awesome!](https://twitter.com/intent/retweet?tweet_id=636497036603428864)

## Usage

Setup consists of:

* Setting up a [module bundler](#module-bundler) to load the [Interoperable CSS](https://github.com/css-modules/icss).
* [Decorating](#decorator) your component using `react-css-modules`.

### Module Bundler

#### webpack

##### Development

In development environment, you want to [Enable Sourcemaps](#enable-sourcemaps) and webpack [Hot Module Replacement](https://webpack.github.io/docs/hot-module-replacement.html) (HMR). [`style-loader`](https://github.com/webpack/style-loader) already supports HMR. Therefore, Hot Module Replacement will work out of the box.

Setup:

* Install [`style-loader`](https://www.npmjs.com/package/style-loader).
* Install [`css-loader`](https://www.npmjs.com/package/css-loader).
* Setup `/\.css$/` loader:

```js
{
    test: /\.css$/,
    loaders: [
        'style?sourceMap',
        'css?modules&importLoaders=1&localIdentName=[path]___[name]__[local]___[hash:base64:5]'
    ]
}
```

##### Production

In production environment, you want to extract chunks of CSS into a single stylesheet file.

> Advantages:
>
> * Fewer style tags (older IE has a limit)
> * CSS SourceMap (with `devtool: "source-map"` and `css-loader?sourceMap`)
> * CSS requested in parallel
> * CSS cached separate
> * Faster runtime (less code and DOM operations)
>
> Caveats:
>
> * Additional HTTP request
> * Longer compilation time
> * More complex configuration
> * No runtime public path modification
> * No Hot Module Replacement

– [extract-text-webpack-plugin](https://github.com/webpack/extract-text-webpack-plugin)

Setup:

* Install [`style-loader`](https://www.npmjs.com/package/style-loader).
* Install [`css-loader`](https://www.npmjs.com/package/css-loader).
* Use [`extract-text-webpack-plugin`](https://www.npmjs.com/package/extract-text-webpack-plugin) to extract chunks of CSS into a single stylesheet.
* Setup `/\.css$/` loader:

```js
{
    test: /\.css$/,
    loader: ExtractTextPlugin.extract('style', 'css?modules&importLoaders=1&localIdentName=[name]__[local]___[hash:base64:5]')
}
```

* Setup `extract-text-webpack-plugin` plugin:

```js
new ExtractTextPlugin('app.css', {
    allChunks: true
})
```

Refer to [webpack-demo](https://github.com/css-modules/webpack-demo) or [react-css-modules-examples](https://github.com/gajus/react-css-modules-examples) for an example of a complete setup.

##### Browserify

Refer to [`css-modulesify`](https://github.com/css-modules/css-modulesify).

### Extending Component Styles

Use `styles` property to overwrite the default component styles.

Explanation using `Table` component:

```js
import React from 'react';
import CSSModules from 'react-css-modules';
import styles from './table.css';

class Table extends React.Component {
    render () {
        return <div styleName='table'>
            <div styleName='row'>
                <div styleName='cell'>A0</div>
                <div styleName='cell'>B0</div>
            </div>
        </div>;
    }
}

export default CSSModules(Table, styles);
```

In this example, `CSSModules` is used to decorate `Table` component using `./table.css` CSS Modules. When `Table` component is rendered, it will use the properties of the `styles` object to construct `className` values.

Using `styles` property you can overwrite the default component `styles` object, e.g.

```js
import customStyles from './table-custom-styles.css';

<Table styles={customStyles} />;
```

[Interoperable CSS](https://github.com/css-modules/icss) can [extend other ICSS](https://github.com/css-modules/css-modules#dependencies). Use this feature to extend default styles, e.g.

```css
/* table-custom-styles.css */
.table {
    composes: table from './table.css';
}

.row {
    composes: row from './table.css';
}

/* .cell {
    composes: cell from './table.css';
} */

.table {
    width: 400px;
}

.cell {
    float: left; width: 154px; background: #eee; padding: 10px; margin: 10px 0 10px 10px;
}
```

In this example, `table-custom-styles.css` selectively extends `table.css` (the default styles of `Table` component).

Refer to the [`UsingStylesProperty` example](https://github.com/gajus/react-css-modules-examples/tree/master/src/UsingStylesProperty) for an example of a working implementation.

### `styles` Property

Decorated components inherit `styles` property that describes the mapping between CSS modules and CSS classes.

```js
class extends React.Component {
    render () {
        <div>
            <p styleName='foo'></p>
            <p className={this.props.styles.foo}></p>
        </div>;
    }
}
```

In the above example, `styleName='foo'` and `className={this.props.styles.foo}` are equivalent.

`styles` property is designed to enable component decoration of [Loops and Child Components](#loops-and-child-components).

### Loops and Child Components

`styleName` cannot be used to define styles of a `ReactElement` that will be generated by another component, e.g.

```js
import React from 'react';
import CSSModules from 'react-css-modules';
import List from './List';
import styles from './table.css';

class CustomList extends React.Component {
    render () {
        let itemTemplate;

        itemTemplate = (name) => {
            return <li styleName='item-template'>{name}</li>;
        };

        return <List itemTemplate={itemTemplate} />;
    }
}

export default CSSModules(CustomList, styles);
```

The above example will not work. `CSSModules` is used to decorate `CustomList` component. However, it is the `List` component that will render `itemTemplate`.

For that purpose, the decorated component inherits [`styles` property](#styles-property) that you can use just as a regular CSS Modules object. The earlier example can be therefore rewritten to:

```js
import React from 'react';
import CSSModules from 'react-css-modules';
import List from './List';
import styles from './table.css';

class CustomList extends React.Component {
    render () {
        let itemTemplate;

        itemTemplate = (name) => {
            return <li className={this.props.styles['item-template']}>{name}</li>;
        };

        return <List itemTemplate={itemTemplate} />;
    }
}

export default CSSModules(CustomList, styles);
```

You can use `styleName` property within the child component if you decorate the child component using `CSSModules` before passing it to the rendering component, e.g.

```js
import React from 'react';
import CSSModules from 'react-css-modules';
import List from './List';
import styles from './table.css';

class CustomList extends React.Component {
    render () {
        let itemTemplate;

        itemTemplate = (name) => {
            return <li styleName='item-template'>{name}</li>;
        };

        itemTemplate = CSSModules(itemTemplate, this.props.styles);

        return <List itemTemplate={itemTemplate} />;
    }
}

export default CSSModules(CustomList, styles);
```

### Decorator

```js
/**
 * @typedef CSSModules~Options
 * @see {@link https://github.com/gajus/react-css-modules#options}
 * @property {Boolean} allowMultiple
 * @property {Boolean} errorWhenNotFound
 */

/**
 * @param {Function} Component
 * @param {Object} defaultStyles CSS Modules class map.
 * @param {CSSModules~Options} options
 * @return {Function}
 */
```

You need to decorate your component using `react-css-modules`, e.g.

```js
import React from 'react';
import CSSModules from 'react-css-modules';
import styles from './table.css';

class Table extends React.Component {
    render () {
        return <div styleName='table'>
            <div styleName='row'>
                <div styleName='cell'>A0</div>
                <div styleName='cell'>B0</div>
            </div>
        </div>;
    }
}

export default CSSModules(Table, styles);
```

Thats it!

As the name implies, `react-css-modules` is compatible with the [ES7 decorators](https://github.com/wycats/javascript-decorators) syntax:

```js
import React from 'react';
import CSSModules from 'react-css-modules';
import styles from './table.css';

@CSSModules(styles)
export default class extends React.Component {
    render () {
        return <div styleName='table'>
            <div styleName='row'>
                <div styleName='cell'>A0</div>
                <div styleName='cell'>B0</div>
            </div>
        </div>;
    }
}
```

[Awesome!](https://twitter.com/intent/retweet?tweet_id=636497036603428864)

Refer to the [react-css-modules-examples](https://github.com/gajus/react-css-modules-examples) repository for an example of webpack setup.

### Options

Options are supplied as the third parameter to the `CSSModules` function.

```js
CSSModules(Component, styles, options);
```

or as a second parameter to the decorator:

```js
@CSSModules(styles, options);
```

#### `allowMultiple`

Default: `false`.

Allows multiple CSS Module names.

When `false`, the following will cause an error:

```js
<div styleName='foo bar' />
```

#### `errorWhenNotFound`

Default: `true`.

Throws an error when `styleName` cannot be mapped to an existing CSS Module.

## SASS, SCSS, LESS and other CSS Preprocessors

[Interoperable CSS](https://github.com/css-modules/icss) is compatible with the CSS preprocessors. To use a preprocessor, all you need to do is add the preprocessor to the chain of loaders, e.g. in the case of webpack it is as simple as installing `sass-loader` and adding `!sass` to the end of the `style-loader` loader query (loaders are processed from right to left):

```js
{
    test: /\.scss$/,
    loaders: [
        'style',
        'css?modules&importLoaders=1&localIdentName=[path]___[name]__[local]___[hash:base64:5]',
        'resolve-url',
        'sass'
    ]
}
```

### Enable Sourcemaps

To enable CSS Source maps, add `sourceMap` parameter to the css-loader and to the `sass-loader`:

```js
{
    test: /\.scss$/,
    loaders: [
        'style?sourceMap',
        'css?modules&importLoaders=1&localIdentName=[path]___[name]__[local]___[hash:base64:5]',
        'resolve-url',
        'sass?sourceMap'
    ]
}
```

## Class Composition

CSS Modules promote composition pattern, i.e. every CSS Module that is used in a component should define all properties required to describe an element, e.g.

```css
.box {
    width: 100px;
    height: 100px;
}

.empty {
    composes: box;

    background: #4CAF50;
}

.full {
    composes: box;

    background: #F44336;
}
```

Composition promotes better separation of markup and style using semantics that would be hard to achieve without CSS Modules.

Because CSS Module names are local, it is perfectly fine to use generic style names such as "empty" or "full", without "box-" prefix.

To learn more about composing CSS rules, I suggest reading Glen Maddern article about [CSS Modules](http://glenmaddern.com/articles/css-modules) and the official [spec of the CSS Modules](https://github.com/css-modules/css-modules).

### What Problems does Class Composition Solve?

Consider the same example in CSS and HTML:

```css
.box {
    width: 100px;
    height: 100px;
}

.box-empty {
    background: #4CAF50;
}

.box-full {
    background: #F44336;
}
```

```html
<div class='box box-empty'></div>
```

This pattern emerged with the advent of OOCSS. The biggest disadvantage of this implementation is that you will need to change HTML almost every time you want to change the style.

### Class Composition Using CSS Preprocessors

This section of the document is included as a learning exercise to broaden the understanding about the origin of Class Composition. CSS Modules support a native method of composting CSS Modules using [`composes`](https://github.com/css-modules/css-modules#composition) keyword. CSS Preprocessor is not required.

You can write compositions in SCSS using [`@extend`](http://sass-lang.com/documentation/file.SASS_REFERENCE.html#extend) keyword and using [Mixin Directives](http://sass-lang.com/documentation/file.SASS_REFERENCE.html#mixins), e.g.

Using `@extend`:

```css
%box {
    width: 100px;
    height: 100px;
}

.box-empty {
    @extend %box;

    background: #4CAF50;
}

.box-full {
    @extend %box;

    background: #F44336;
}
```

This translates to:

```css
.box-empty,
.box-full {
    width: 100px;
    height: 100px;
}

.box-empty {
    background: #4CAF50;
}

.box-full {
    background: #F44336;
}
```

Using mixins:

```css
@mixin box {
    width: 100px;
    height: 100px;
}

.box-empty {
    @include box;

    background: #4CAF50;
}

.box-full {
    @include box;

    background: #F44336;
}
```

This translates to:

```css
.box-empty {
    width: 100px;
    height: 100px;
    background: #4CAF50;
}

.box-full {
    width: 100px;
    height: 100px;
    background: #F44336;
}
```

## Global CSS

CSS Modules does not restrict you from using global CSS.

```css
:global .foo {

}
```

However, use global CSS with caution. With CSS Modules, there are only a handful of valid use cases for global CSS (e.g. [normalization](https://github.com/necolas/normalize.css/)).

## Multiple CSS Modules

Avoid using multiple CSS Modules to describe a single element. Read about [Class Composition](#class-composition).

That said, if you require to use multiple CSS Modules to describe an element, enable the [`allowMultiple`](#allowmultiple) option. When multiple CSS Modules are used to describe an element, `react-css-modules` will append a unique class name for every CSS Module it matches in the `styleName` declaration, e.g.

```css
.button {

}

.active {

}
```

```js
<div styleName='button active'></div>
```

This will map both [Interoperable CSS](https://github.com/css-modules/icss) CSS classes to the target element.
