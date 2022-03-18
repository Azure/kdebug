# kdebug

kdebug is a command line utility that helps troubleshoot a running Kubernetes cluster and apps in it.

By running a set of predefined checkers, it gives diagnostics information and guides you to next steps.

## How to use

Run all suites:

```bash
kdebug
```

Run a specific suites:

```bash
kdebug -s dns
```

List available suites:

```bash
kdebug --list
```

Batch mode:

```
kdebug -s dns \
    --batch.machines=machine-1 \
    --batch.machines=machine-2 \
    --batch.concurrency=2 \
    --batch.sshuser=azureuser
```

## Development

Prerequisite:

* [Golang](https://go.dev/dl/)

Build:

```bash
make build
```

Test:

```bash
make test
```

## Contributing

This project welcomes contributions and suggestions.  Most contributions require you to agree to a
Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us
the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide
a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions
provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/).
For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or
contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft
trademarks or logos is subject to and must follow
[Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general).
Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship.
Any use of third-party trademarks or logos are subject to those third-party's policies.
