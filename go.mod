module github.com/redradrat/shipcaps

go 1.13

require (
	github.com/Masterminds/sprig v0.0.0-20190301161902-9f8fceff796f
	github.com/aws/aws-sdk-go v1.27.4
	github.com/fluxcd/helm-operator v1.0.0-rc6
	github.com/go-logr/logr v0.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.7.0
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.0.0-20191114100352-16d7abae0d2a
	k8s.io/apimachinery v0.0.0-20191028221656-72ed19daf4bb
	k8s.io/client-go v12.0.0+incompatible
	k8s.io/utils v0.0.0-20191114184206-e782cd3c129f
	sigs.k8s.io/controller-runtime v0.4.0
)

replace github.com/docker/distribution => github.com/2opremio/distribution v0.0.0-20200223014041-6b972e50feee

replace github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90
