# Storm

This is a living document that we will be updating as we work on the Storm project. It documents patterns we want to deprecate in the components migrated under Storm routes. This draft version will eventually be moved to the documentation website once we calibrate them on more real-world examples.

## Guidelines

1. Interfaces or types extension: prefer explicitly listing all the fields in the interface or object or using `Pick<MyObject, 'fieldName'>` to select required fields to make interfaces readable and explicit. If this recommendation adds too much overhead, it probably means that a different approach should be taken for sharing data between components. There should be no reason to reuse the same interface with multiple fields on many React tree levels. Use React.Context or Apollo Client or zustand stores to bypass component layers.

```tsx
// Good
interface MyComponentProps extends Pick<MyContext, 'sharedField'> {
  email: string
  isFeatureEnabled: true
  children: ReactNode
}

// Bad
interface MyComponentProps extends MyContext, AuthenticatedUser, FeatureContext {
  children: ReactNode
}
```

2. Use Apollo Client for fetching data from our GraphQL API.

3. Do not use `rxjs` unless there's no other way to achieve the same result. Prefer Apollo Client, React hooks, zustand stores, or whatever else you can find. `rxjs` should be used as a last resort to solve complex concurrency issues.

4. Do not use `useMemo` or `useCallback` by default. Use them to solve the concrete performance bottleneck. See https://kentcdodds.com/blog/usememo-and-usecallback for more context and examples.

5. Do not rename variables, imports, fields, etc. Always prefer the same name as in the source. It makes it easier to find things across the codebase and follow the logic in components.
