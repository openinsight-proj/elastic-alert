# Elastic-Alert Helm Charts

This functionality is in beta and is subject to change. 
The code is provided as-is with no warranties. Beta features are not subject to the support SLA of official GA features.

## Usage

[Helm](https://helm.sh/) must be installed to use the charts. Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```bash
$ helm repo add openinsight-proj https://openinsightio.github.io/elastic-alert
```

You can then run `helm search repo openinsight-proj` to see the charts.

And then, install elastic-alert by:

```bash
$ helm upgrade --install elastic-alert openinsight-proj/elastic-alert
```

## Contributing
We'd love to have you contribute! Please refer to our [contribution guidelines](https://github.com/openinsight-proj/elastic-alert/blob/main/CONTRIBUTING.md) for details.

## License

Apache 2.0 License.
