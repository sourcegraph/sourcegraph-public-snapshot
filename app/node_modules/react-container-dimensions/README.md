# react-container-dimensions [![Build Status](https://travis-ci.org/okonet/react-container-dimensions.svg?branch=master)](https://travis-ci.org/okonet/react-container-dimensions)
Wrapper component that detects parent (container) element resize and passes new dimensions down the 
tree. Based on [element-resize-detector]
(https://github.com/wnr/element-resize-detector)

`npm install --save react-container-dimensions`

It is especially useful when you create components with dimensions that change over 
time and you want to explicitely pass the container dimensions to the children. For example, SVG 
visualization needs to be updated in order to fit into container.

## Usage

* Wrap your existing components. Children component will recieve `width` and `height` as props. 

```jsx
<ContainerDimensions>
    <MyComponent/>
</ContainerDimensions>    
```

* Use a function to pass width or height explicitely or do some calculation. Function callback will be called with an object `{ width: number, height: number }` as an argument and it expects the output to be a React Component or an element. 

```jsx
<ContainerDimensions>
    { ({ height }) => <MyComponent height={height}/> }
</ContainerDimensions>    
```

## How is it different from ...

*It does not create a new element in the DOM but relies on the `parentNode` which must be present.* 
This means it doesn't require its own CSS to do the job and leaves it up to you. So, basically, 
it acts as a middleware to pass _your_ styled component dimensions to your children components. This makes it _very_ easy to integrate with your existing code base.

## Example

Let's say you want your SVG visualization to always fit into the container. In order for SVG to scale elements properly it is required that `width` and `height` attributes are properly set on the `svg` element. Imagine the following example

### Before (static)

It's hard to keep dimensions of the container and the SVG in sync. Especially, when you want your content to be resplonsive (or dynamic).

```jsx
export const myVis = () => (
    <div className="myStyles">
        <svg width={600} height={400}>
            {/* SVG contents */}
        </svg>  
    <div>
)
```

### After (dynamic)

This will resize and re-render the SVG each time the `div` dimensions are changed. For instance, when you change CSS for `.myStyles`.

```jsx
import ContainerDimensions from 'react-container-dimensions'

export const myVis = () => (
    <div className="myStyles">
        <ContainerDimensions>
            { ({ width, height }) => 
                <svg width={width} height={height}>
                    {/* SVG contents */}
                </svg>  
            }
        </ContainerDimensions>
    <div>
)
```

## Other similar projects:

* https://github.com/maslianok/react-resize-detector
* https://github.com/Xananax/react-size
* https://github.com/joeybaker/react-element-query

and a few others...
