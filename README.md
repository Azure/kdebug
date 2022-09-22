# kdebug

kdebug is a command line utility that helps troubleshoot a running Kubernetes cluster and apps in it.

It focuses on DevOps scenarios and covers these areas:

* OS diagnostics
* Kubernetes components diagnostics
* Lightweight application diagnostics

## Check mode

kdebug runs in check mode by default.
By running a set of predefined checks, it gives diagnostics information and guides you to next steps.

Currently kdebug supports following checks:

* Disk usage: Check disk usage and identity top large files.
* DNS: Check cluster DNS.
* HTTP: Check HTTP connectivity to well known endpoints.
* Kube Object Size: Check configmap/secret object size.
* Kube pod: Check pod restart reasons.
* OOM: Analysis out-of-memory events.
* System Load: Check the CPU and Memory of VM and some primary processes (etcd, kubelet...)

## How to use

### Basic

Run all checks:

```bash
kdebug
```

Run a specific check:

```bash
kdebug -c dns
```

List available checks:

```bash
kdebug --list
```

See full supported arguments and help:

```bash
kdebug -h
```

### Kubernetes checks

Kubernetes related checks require a working kubeconfig. You can either put it at the default location `$HOME/.kube/config`, or you can specify via `--kube-config-path`:

```bash
kdebug -c kubepod \
    --kube-config-path /path/to/kubeconfig
```

### Batch mode

kdebug supports running on a batch of remote machines simultaneously via SSH.

Explictly specify a list of machine names:

```bash
kdebug -c dns \
    --batch.machines=machine-1 \
    --batch.machines=machine-2 \
    --batch.concurrency=2 \
    --batch.ssh-user=azureuser
```

Read machine names list from a file or stdin:

```bash
# From file
kdebug -c dns \
    --batch.machines-file=/path/to/machine/names/file

# From stdin
kubectl get nodes | grep NotReady | awk '{print $1}' | kdebug -c dns --batch.machines-file=-
```

Auto discover machines list via Kubernetes API server.

```bash
kdebug -c dns --batch.kube-machines
```

In addition, you can specify a label selector:

```bash
kdebug -c dns \
    --batch.kube-machines \
    --batch.kube-machines-label=kubernetes.io/role=agent
```

Or filter out unready nodes only:

```bash
kdebug -c dns \
    --batch.kube-machines-unready
```

## Tool mode

In addition to the default check mode, kdebug also supports a tool mode.
Tool mode wraps useful commands and makes them easier to used in typical scenarios.


Currently kdebug provides following tools:

* Tcpdump: Wrap tcpdump command and provides a simpler interface for container scenarios.
* Reboot reason: Inspect last reboot reason.
* AAD SSH: SSH via AAD. This is a handy replacement for the original Azure CLI based implementation.
* NetExec: Execute the command with the same network namespace with a specific process or pod.

You can see a full list with:

```bash
kdebug --list
```

Use following command to start a tool:

```bash
kdebug -t <tool>
```

Show tool specific options:

```bash
kdebug -t <tool> -h
```

### Tcpdump

Attach to network namespace of a process with pid=100 and capture all traffic:

```bash
kdebug -t tcpdump --pid=100
```

With source and destination specified, and TCP only:

```bash
kdebug -t tcpdump \
    --pid=100 \
    --source=10.0.0.1:1000 \
    --destination=10.0.0.2:2000 \
    --tcponly
```

`--host` matches either source or destination:

```bash
kdebug -t tcpdump --host=10.0.0.1:1000
```

### Reboot reason

Check VM last reboot reason within last 1 day:

```
kdebug -t vmrebootdetector
```

Check VM last reboot reason within last 100 days:

```
kdebug -t vmrebootdetector \
    --checkdays=100
```

### Package upgrade inspect

Check upgraded packages within last 14 days:

```
kdebug --tool upgradeinspector --checkdays 14
```

Check upgraded package within last 7 days, limit 10 records:

```
kdebug --tool upgradeinspector --recordlimit 10
```

### AAD SSH

SSH via AAD. See [Azure Linux VMs and Azure AD](https://learn.microsoft.com/en-us/azure/active-directory/devices/howto-vm-sign-in-azure-ad-linux).

This is a handy replacement for the original Azure CLI based implementation.

Login via interactive flow:

```bash
kdebug -t aadssh <user>@<tenant>@<hostname-or-ip>
```

A browser will pop up for credentials.

Login via Azure CLI credentials:

```bash
az login
kdebug -t aadssh --use-azure-cli <user>@<tenant>@<hostname-or-ip>
```

### NetExec
Execute the command with the same network namespace with a process, you need to on the VM the process locate in.

```bash
kdebug -t netexec --pid=<process-pid>
```

Execute the command with the same network namespace with a pod, you need to have the kubeconfig.

```bash
kdebug -t netexec --pod=<pod-name> --namespace=<pod-namespace>
```

And specify the command with `--command=`. The default command is `sh`

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
