# Shipcaps [![Join the chat at https://gitter.im/shipcaps/community](https://badges.gitter.im/shipcaps/community.svg)](https://gitter.im/shipcaps/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

<p align="center">
	<img src="logo.png" width="20%" align="center" alt="ExternalDNS">
</p>

A meta layer for kubernetes applications. Shipcaps provides a meta layer over various awsome packaging tools and projects. 
With our Custom Resources, we represent *providing* a kubernetes application, and *using* a kubernetes application in a 
normalized fashion. We embrace kubernetes' ecosystem diversity and are normalizing operations with various existing 
tools and do not aim to replace them.

Table of Contents
=================

  * [Idea](#idea)
  * [Design](#design)
     * [Cap ("Capability")](#cap-capability)
        * [Material](#material)
           * [repo](#repo)
           * [manifests](#manifests)
        * [Types](#types)
           * [simple](#simple)
           * [helmchart](#helmchart)
     * [App ("Application")](#app-application)

## Idea

Maintaining a kubernetes offering for any type of organisation can be a challange. In order to minimize the knowledge 
required to run a service on kubernetes, a couple of tools emerge(d) to solve the problem of packaging Applications.
Those packages consist of a number of kubernetes native resources, and are handled as a whole via Helm, Kustomize, Ship, 
ManifestBundle, CosmicEgg or FluxCapacitor, in their various specific ways.

Kubernetes Engineers/SREs/DevOpsEngineers usually are trained to deal with those formats the same way a Golang Developer 
knows how to handle `go.mod`, `Gopkg.toml` and `glide.yaml`, or a Java Developer knows how to tame `pom.xml` and
`build.gradle`. So we know how to come up with `values.yaml`, `kustomize.yaml`, etc. but sometimes also have the need to
interface with our peers on our platforms that might not have this skillset.

**The idea for *Shipcaps* is to separate the development, tuning and providing of kubernetes-native Applications
and the usage/comsumption of packaged kubernetes applications by users in a normalized fashion.** 

Scenario 1:
In Acme, Corp. I want to provide a common platform for services for my product. Can I provide teams with a common way to
get their dependencies? Do we need to get compliance/security clearance if Microservice A uses acme/postgres-chart and 
Mircoservice B a postgres operator?

Scenario 2:
In Acme, Corp. we want to cut resource costs. As an SRE I'm currently going through all deployments and am streamlining 
resource issues and autoscaling. I can only do that because I am efficient using 4 different packaging tools. Could this
streamlining have been done by the users themselves? 

Scenario 3:
In Acme, Corp. we want a constant overview over deployments for compliance and security reasons. How does this inventory
look like? Can I just list all currently running applications, and inspect their parts if an audit comes? GitOps maybe? 

Scenario 4:
In Acme, Corp. we want to allocate team's budgets based on whether they provide value for other teams. Team A plans to 
create a custom postgres chart. Team B, C, D and E need postgres too. Do I have to pay for development if I'm managing 
Team A? Can I relay development costs to thos other teams somehow? 

Scenario 5:
In Acme, Corp. we want to have one or more repositories of kubernetes-native applications, goverened by a distributed
body. Do I have a corporate-wide registry of them? I might have a couple of helm chart registries, and a couple of 
common kustomize repos. But utilizing it, I still need some domain knowledge to utilize and explore those. Is there a 
technical representation for this, that I can act upon and inventorize?

**Conclusion**

Those are **awesome** scenarios already. We're getting stuff done in potentially cross-functioning teams, and we're 
packaging clusters of kubernetes resources into logical units, We're even going towards having a single source of truth 
with little magic between commit and deployment! Arriving here means you're probably already doing loads of cool stuff.

## Design

In Shipcaps we're dealing with 2 different kinds. A `Cap` as in "Capability", and an `App` as in "Application".

![design](idea-sketch.png)

### Cap ("Capability")

See [examples/helmcap.yaml](./examples/helmcap.yaml).

A `Cap` defines a packaged kubernetes application. This can be a couple of manifests with maybe a couple of 
placeholder-values for environment-specific config, or a Helm Chart checked into some common git repository I want to 
feed some custom values to, or something entirely different.

With a `Cap` you can refer to a package and define what inputs it needs.

#### Material

Each Cap refers to underlying material. Again, this project does not aim to replace existing tooling but only be an 
abstraction of "value in, application out" usecases.

Usage:
```yaml
apiVersion: shipcaps.redradrat.xyz/v1beta1
kind: Cap
metadata:
  name: acme-es
spec:
  material: ## Here goes our material spec
```


**Supported Material**:

##### repo

The Git Repo material is really the embodyment of our GitOps strategy. With this material spec we can refer to an 
existing git repo or a subdirectory in it, and expect our [type](#types)-specific sources at this location.

Usage:
```yaml
...
spec:
  ...
  material:
    repo:
      uri: https://github.com/redradrat/charts
      ref: v1.0
      auth:
        username:
          secretKeyRef:
            name: repoauth
            key: repo-username
        password: 
          secretKeyRef:
            name: repoauth
            key: repo-password
    path: /elasticsearch
```

##### manifests

The manifests material spec is a quick-and-easy way to abstract a single or a couple of manifests into a Cap.

Usage:
```yaml
...
spec:
  material:
    ...
    manifests:
      - apiVersion: v1
        kind: Namespace
        metadata:
        name: es
      - apiVersion: elasticsearch.k8s.elastic.co/v1
        kind: Elasticsearch
        metadata:
          name: my-elasticsearch
          namespace: es
        spec:
          version: 7.6.2
          nodeSets:
          - name: default
            count: 1
          ...
```

#### Types

There are various different *types* of Caps. They all represent different ways of templating, generating or otherwise 
compinling manifests or input towards the kube-apiserver. With caps we just want to come up with a list of values to 
pass on to our "backend", the actual tool used to generate.

Usage:
```yaml
apiVersion: shipcaps.redradrat.xyz/v1beta1
kind: Cap
metadata:
  name: acme-es
spec:
  type: __TYPE_GOES_HERE__
  inputs:
    - key: mykey
      ...
```


**Supported types**:

##### simple

The `simple` Cap type refers to plain kubernetes manifests enriched with simple placeholder-replacement functionality.

Usecases:
* Single or few ready-made manifests, to be applied to various environments. (Domain name, varying)
* Operator Deployment coupled with CRDs
* ...

Supported material:
* repo
* manifests

##### helmchart

The `helmchart` Cap type refers to a helm chart and allows to defines a set of inputs. This type expects the the 
[helm-operator](https://github.com/fluxcd/helm-operator/) to be available.

Supported material:
* repo


### App ("Application")

See [examples/postgresapp.yaml](./examples/postgresapp.yaml)

An `App` defines an instance of a `Cap`. It references the `Cap` and defines the values it requires. After being 
reconciled by the shipcaps operator, the application will be usable.

Usage:
```yaml
apiVersion: shipcaps.redradrat.xyz/v1beta1
kind: App
metadata:
  name: myelastic
  namespace: here
spec:
  capRef: acme-es # This is a single string, as caps are cluster-wide
  values:
    - key: dbname
      value: mydb

```
