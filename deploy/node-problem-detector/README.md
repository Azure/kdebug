# What is npd

node-problem-detector aims to make various node problems visible to the upstream layers in the cluster management stack. It is a daemon that runs on each node, detects node problems and reports them to apiserver. node-problem-detector can either run as a DaemonSet or run standalone. Now it is running as a Kubernetes Addon enabled by default in the GCE cluster.

# How to deploy npd

In this project, we integrate the node-problem-detector with kdebug. You can run the following command to deploy the integrated daemon app to your kubernetes cluster.
```shell
kubectl deploy -f ./node-problem-detector/node-problem-detector.yaml
```

# What can you find with npd