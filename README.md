# Shipcaps

A meta layer for kubernetes applications. Shipcaps provides a meta layer over various awsome packaging tools and projects. 
With our Custom Resources, we represent *providing* a kubernetes application, and *using* a kubernetes application in a 
normalized fashion. We embrace kubernetes' ecosystem diversity and are normalizing operations with various existing 
tools and do not aim to replace them.

## Design

In Shipcaps we're dealing with 2 different types. A `Cap` as in "Capability", and an `App` as in "Application".

### Cap ("Capability")

See examples/helmcap.yaml.

A `Cap` defines a packaged kubernetes application. This can be a couple of manifests with maybe a couple of 
placeholder-values for environment-specific config, or a Helm Chart checked into some common git repository I want to 
feed some custom values to, or something entirely different.

With a `Cap` you can refer to a package and define what inputs it needs.

### App ("Application")

See examples/postgresapp.yaml

An `App` defines an instance of a `Cap`. It specifies the `Cap` and the values it needs. After being reconciled by the 
shipcaps operator, the application will be usable.

## Idea

Maintaining a kubernetes offering for any type of organisation can be a challange. In order to minimize the knowledge 
required to run a service on kubernetes, a couple of tools emerge(d) to solve the problem of packaging Applications.
Thos packages consist of a number of kubernetes native resources, and are handled as a whole via Helm, Kustomize, Ship, 
Manifest Bundle, Cosmic Egg or FluxCapacitor, in their various specific was.

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

Those are **awesome** scenarios already. We're getting stuff done in potentially cross-functioning teams, and we're 
packaging clusters of kubernetes resources into logical units, We're even going towards having a single source of truth 
with little magic between commit and deployment! Arriving here means you're probably already doing loads of cool stuff.

Now the point is, what are those units? Are they all kubernetes applications? Do I have a corporate-wide registry of 
them? I might have a couple of helm chart registries, and a couple of common kustomize repos. But utilizing it, I still
need some domain knowledge. Also which cost center pays for those efforts? Can I cross-charge development efforts to 
departments that are using it? How could I measure that? Shipcaps provides this API.
