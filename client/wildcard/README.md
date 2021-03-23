# Wildcard Component Library

## Overview

The Wildcard component library is a collection of design-approved reusable components that are suitable for use within the Sourcegraph codebase.

## Usage

This library is setup as a `yarn workspace` symlink.

You can import components from the library like so:

```javascript
import { PageSelector } from '@sourcegraph/wildcard'
```

## Structure

- `src/`
  - `components/`
    Reusable React components
  - `hooks/`
    Reusable React hooks. Typically utilities or headless components
