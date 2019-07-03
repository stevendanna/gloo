
---
title: "kubernetes.proto"
weight: 5
---

<!-- Code generated by solo-kit. DO NOT EDIT. -->


### Package: `kubernetes.plugins.gloo.solo.io` 
#### Types:


- [UpstreamSpec](#upstreamspec)
  



##### Source File: [github.com/solo-io/gloo/projects/gloo/api/v1/plugins/kubernetes/kubernetes.proto](https://github.com/solo-io/gloo/blob/master/projects/gloo/api/v1/plugins/kubernetes/kubernetes.proto)





---
### UpstreamSpec

 
Upstream Spec for Kubernetes Upstreams
Kubernetes Upstreams represent a set of one or more addressable pods for a Kubernetes Service
the Gloo Kubernetes Upstream maps to a single service port. Because Kubernetes Services support multiple ports,
Gloo requires that a different upstream be created for each port
Kubernetes Upstreams are typically generated automatically by Gloo from the Kubernetes API

```yaml
"serviceName": string
"serviceNamespace": string
"servicePort": int
"selector": map<string, string>
"serviceSpec": .plugins.gloo.solo.io.ServiceSpec
"subsetSpec": .plugins.gloo.solo.io.SubsetSpec

```

| Field | Type | Description | Default |
| ----- | ---- | ----------- |----------- | 
| `serviceName` | `string` | The name of the Kubernetes Service |  |
| `serviceNamespace` | `string` | The namespace where the Service lives |  |
| `servicePort` | `int` | The access port of the kubernetes service is listening. This port is used by Gloo to look up the corresponding port on the pod for routing. |  |
| `selector` | `map<string, string>` | Allows finer-grained filtering of pods for the Upstream. Gloo will select pods based on their labels if any are provided here. (see [Kubernetes labels and selectors](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/) |  |
| `serviceSpec` | [.plugins.gloo.solo.io.ServiceSpec](../../service_spec.proto.sk#servicespec) | An optional Service Spec describing the service listening at this address |  |
| `subsetSpec` | [.plugins.gloo.solo.io.SubsetSpec](../../subset_spec.proto.sk#subsetspec) | Subset configuration. For discovery sources that has labels (like kubernetes). this configuration allows you to partition the upstream to a set of subsets. for each unique set of keys and values, a subset will be created. |  |





<!-- Start of HubSpot Embed Code -->
<script type="text/javascript" id="hs-script-loader" async defer src="//js.hs-scripts.com/5130874.js"></script>
<!-- End of HubSpot Embed Code -->