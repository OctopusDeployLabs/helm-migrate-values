# helm-migrate-values

A plugin to migrate user-specified Helm values between chart versions when the schema of `values.yaml` changes. Define migration paths with migration files in the chart repository to ensure seamless upgrades.
## Requirements

- **Helm** version 3 or newer
- **Git** if installing using the `helm plugin install` command below

## Install

```
$ helm plugin install https://github.com/OctopusDeploy/helm-migrate-values.git
```

## Usage

This Helm plugin enables users to migrate values across chart versions, accounting for changes in the values.yaml schema.

> **_NOTE:_** The intended usage of this plugin is that it does not apply any changes to the release on its own. It simply outputs the values to override, which are to be used with `helm upgrade --reset-then-reuse-values`.
> 
> See [output values](#optional-output-the-migration-to-a-file) section for an example on how to apply the migration to a release 

### Step 1: Define the Migration Files
Start by defining the migration files within your Helm chart. These files should be placed under the `value-migrations/` directory, relative to your base chart directory. You can customize the migration directory location by using the `--migration-dir` flag if necessary.

#### Migration File Naming Convention
Each migration file should conform to the following naming format:
`to-v{VERSION_TO}.yaml`, where **VERSION_TO** is the target chart's major version number. The plugin will use this file to define the transformation between the previous version and the specified version.

#### Migration File Structure
Migration files are written in YAML and use Go templating, similar to Helm templates. They leverage Sprig v3's [TxtFuncMap](https://github.com/Masterminds/sprig/blob/fc7fc0d6a0377bca7049c4a99e80b85f222d8caf/functions.go#L49) functions for transforming and mapping values between old and new schemas. See this [example](pkg/test-charts/v2/value-migrations/to-v2.yaml) of a migration definition from the integration test.

### Step 2: Run the Migration
To migrate your Helm release to a new chart version, use the following command:
```
helm migrate-values [RELEASE] [CHART] [flags]
```

- **RELEASE**: The name of the Helm release you're migrating.
- **CHART**: The chart you're migrating to, which can be a local chart (specified by file path) or a remote chart (using the `oci://` or `https://` prefixes).

## Example
```
helm migrate-values my-kubernetes-agent oci://registry-1.docker.io/octopusdeploy/kubernetes-agent \
  --version 2.4.0 \
  -n octopus-agent-demo \
  --migration-dir kubernetes-agent/value-migrations
  --output-file migrated-values.yaml
```

The `--output-file` flag allows you to optionally save the command's output to a file instead of displaying it in stdout, allow you to utilize it in subsequent Helm commands. 

You can then use this file with the helm upgrade command to complete the migration:

```
helm upgrade [RELEASE] [CHART] -f migrated-values.yaml --reset-then-reuse-values
```

## Contributing
We adhere to [Semantic Versioning](https://semver.org/) and utilize [@changesets/cli](https://github.com/changesets/changesets) to manage release notes and versioning.

To add a changelog entry for your changes:

1. Run the following command: `npm run changeset`
2. Select the appropriate version type for your changes (patch, minor, or major).
3. Provide a brief description of the changes you've made.

This will generate a changeset file, which must be included in your pull request.