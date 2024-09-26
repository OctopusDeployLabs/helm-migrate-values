# helm-migrate-values

A plugin to migrate user-specified Helm values between chart versions when the schema of `values.yaml` changes. Define migration paths with migration files in the chart repo to ensure seamless upgrades.
## Requirements

- **Helm** version 3 or newer
- **Git** as the `helm plugin install` command requires it

## Install

```
$ helm plugin install https://github.com/OctopusDeploy/helm-migrate-values.git
```


## Contributing
We adhere to [Semantic Versioning](https://semver.org/) and utilize [@changesets/cli](https://github.com/changesets/changesets) to manage release notes and versioning.

To add a changelog entry for your changes:

1. Run the following command: `npm run changeset`
2. Select the appropriate version type for your changes (patch, minor, or major).
3. Provide a brief description of the changes you've made.

This will generate a changeset file, which must be included in your pull request.