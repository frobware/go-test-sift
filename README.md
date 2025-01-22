# go-test-sift

When Go runs unit tests in parallel, their output becomes an interleaved mess, making debugging test failures difficult. `go-test-sift` restores order by regrouping all output for each test and its subtests into a clean, hierarchical structure, preserving the parent-child relationships.

## Installation

```sh
go install github.com/frobware/go-test-sift@latest
```

## Usage

The simplest case is to pipe Go test output directly to `go-test-sift`:

```sh
go test ./... -v | go-test-sift
```

This will regroup all the interleaved parallel test output into a clean, hierarchical format where each test's output is kept together.

You can also process existing log files or URLs:

```sh
go-test-sift test.log
go-test-sift https://path/to/test.log
```

### Additional Options

To just see test failures:
```sh
go-test-sift -l test.log     # Shows failed test names
go-test-sift -L test.log     # Shows failed tests with their output
```

To save regrouped output to files:
```sh
go-test-sift -w test.log     # Creates directory structure by test name
```

### Test Filtering

The `-t` flag accepts a regular expression to filter which tests to process. It can be used:
- On its own to filter the default regrouped output
- Combined with `-l` or `-L` to filter which failures to show
- Combined with `-w` to filter which test outputs to write to files

Examples:
```sh
# Only show output for specific tests
go-test-sift -t "TestAuth.*" test.log

# Only summarise failures for specific tests
go-test-sift -t "TestAuth.*" -l test.log

# Only write specific test outputs to files
go-test-sift -t "TestAuth.*" -w test.log
```

### Synopsis

```sh
go-test-sift [options] [file|url ...]
  -F	Force directory creation even if directories exist
  -L	Print summary of failures and include the full output for each failure
  -d	Enable debug output
  -l	Print summary of failures (list test names with failures)
  -o string
        Base directory to write output files (default current directory) (default ".")
  -t string
        Regular expression to filter test names for summary output (default ".*")
  -w	Write each test's output to individual files
```

## Real-world usage:

```sh
go-test-sift -L https://storage.googleapis.com/test-platform-results/pr-logs/pull/openshift_cluster-ingress-operator/1182/pull-ci-openshift-cluster-ingress-operator-master-e2e-aws-operator-techpreview/1881996429715574784/build-log.txt
--- FAIL: TestAll (3300.90s)
    --- FAIL: TestAll/parallel (98.74s)
        --- FAIL: TestAll/parallel/Test_IdleConnectionTerminationPolicyDeferred (195.13s)
                idle_connection_test.go:598: Creating namespace "idle-connection-close-deferred-t2mgq"...
                idle_connection_test.go:598: Waiting for ServiceAccount idle-connection-close-deferred-t2mgq/default to be provisioned...
                idle_connection_test.go:598: Waiting for RoleBinding idle-connection-close-deferred-t2mgq/system:image-pullers to be created...
                idle_connection_test.go:598: Creating IngressController openshift-ingress-operator/idle-connection-close-deferred-t2mgq...
                <snip>
                util_test.go:941: 2025-01-22 10:36:27 +0000 UTC {kubelet ip-10-0-124-179.us-east-2.compute.internal} Pod web-service-1 Started Started container web-service-1
                util_test.go:941: 0001-01-01 00:00:00 +0000 UTC { } Pod web-service-2 Scheduled Successfully assigned idle-connection-close-deferred-t2mgq/web-service-2 to ip-10-0-94-112.us-east-2.compute.internal
                util_test.go:941: 2025-01-22 10:36:33 +0000 UTC {multus } Pod web-service-2 AddedInterface Add eth0 [10.129.2.30/23] from ovn-kubernetes
                util_test.go:941: 2025-01-22 10:36:33 +0000 UTC {kubelet ip-10-0-94-112.us-east-2.compute.internal} Pod web-service-2 Pulled Container image "registry.build06.ci.openshift.org/ci-op-q14lpbwi/stable@sha256:e1fa3c38a5cea8a45d127aecd900299ede0c5988495dc6c5db0c1875e4ab4995" already present on machine
                util_test.go:941: 2025-01-22 10:36:33 +0000 UTC {kubelet ip-10-0-94-112.us-east-2.compute.internal} Pod web-service-2 Created Created container web-service-2
                util_test.go:941: 2025-01-22 10:36:33 +0000 UTC {kubelet ip-10-0-94-112.us-east-2.compute.internal} Pod web-service-2 Started Started container web-service-2
                util_test.go:943: Deleting namespace "idle-connection-close-deferred-t2mgq"...
```

```sh
go-test-sift -t Test_IdleConnectionTerminationPolicy https://storage.googleapis.com/test-platform-results/pr-logs/pull/openshift_cluster-ingress-operator/1182/pull-ci-openshift-cluster-ingress-operator-master-e2e-aws-operator-techpreview/1881996429715574784/build-log.txt
        --- PASS: TestAll/parallel/Test_IdleConnectionTerminationPolicyImmediate (162.97s)
                idle_connection_test.go:545: Creating namespace "idle-connection-close-immediate-f5ms9"...
                idle_connection_test.go:545: Waiting for ServiceAccount idle-connection-close-immediate-f5ms9/default to be provisioned...
                idle_connection_test.go:545: Waiting for RoleBinding idle-connection-close-immediate-f5ms9/system:image-pullers to be created...
                idle_connection_test.go:545: Creating IngressController openshift-ingress-operator/idle-connection-close-immediate-f5ms9...
                util_test.go:694: waiting for loadbalancer domain a4420035d95134ad48d26e05284f6ece-18856627.us-east-2.elb.amazonaws.com to resolve...
                util_test.go:694: waiting for loadbalancer domain a4420035d95134ad48d26e05284f6ece-18856627.us-east-2.elb.amazonaws.com to resolve...
                <snip>
                idle_connection_test.go:545: step 3: Ensure subsequent responses are served by web-service-2
                idle_connection_test.go:566: [10.131.224.239:34016 -> 3.22.28.49:80] Req: URL=http://3.22.28.49, Host=test-idle-connection-close-immediate-f5ms9.apps.ci-op-q14lpbwi-9e7c5.origin-ci-int-aws.dev.rhcloud.com
                idle_connection_test.go:566: [10.131.224.239:34016 <- 3.22.28.49:80] Res: Status=200, Headers=map[Content-Length:[8] Content-Type:[text/plain; charset=utf-8] Date:[Wed, 22 Jan 2025 10:37:07 GMT] Set-Cookie:[6d90e534fc735aa6806d7332b8f3e32b=e5ca86e727e8ef81afc42bbcdaa1a897; path=/; HttpOnly] X-Pod-Name:[web-service-2] X-Pod-Namespace:[unknown-namespace]]
                idle_connection_test.go:397: deleted ingresscontroller idle-connection-close-immediate-f5ms9
                util_test.go:939: Dumping events in namespace "idle-connection-close-immediate-f5ms9"...
                util_test.go:943: Deleting namespace "idle-connection-close-immediate-f5ms9"...
        --- FAIL: TestAll/parallel/Test_IdleConnectionTerminationPolicyDeferred (195.13s)
                idle_connection_test.go:598: Creating namespace "idle-connection-close-deferred-t2mgq"...
                idle_connection_test.go:598: Waiting for ServiceAccount idle-connection-close-deferred-t2mgq/default to be provisioned...
                idle_connection_test.go:598: Waiting for RoleBinding idle-connection-close-deferred-t2mgq/system:image-pullers to be created...
                idle_connection_test.go:598: Creating IngressController openshift-ingress-operator/idle-connection-close-deferred-t2mgq...
                util_test.go:694: waiting for loadbalancer domain a9eeec12dfe994550aa9fd1687e66b65-1711296474.us-east-2.elb.amazonaws.com to resolve...
                util_test.go:694: waiting for loadbalancer domain a9eeec12dfe994550aa9fd1687e66b65-1711296474.us-east-2.elb.amazonaws.com to resolve...
                <snip>
                idle_connection_test.go:598: step 1: Verify the initial response is correctly served by web-service-1
                idle_connection_test.go:602: [10.131.224.239:43382 -> 3.128.128.128:80] Req: URL=http://3.128.128.128, Host=test-idle-connection-close-deferred-t2mgq.apps.ci-op-q14lpbwi-9e7c5.origin-ci-int-aws.dev.rhcloud.com
                idle_connection_test.go:602: [10.131.224.239:43382 <- 3.128.128.128:80] Res: Status=200, Headers=map[Content-Length:[8] Content-Type:[text/plain; charset=utf-8] Date:[Wed, 22 Jan 2025 10:36:56 GMT] Set-Cookie:[357f8d00002ad5a815d58dad294244fc=a63bc34bc547085dc2dbe253c0550c80; path=/; HttpOnly] X-Pod-Name:[web-service-1] X-Pod-Namespace:[unknown-namespace]]
                idle_connection_test.go:598: step 2: Switch route to web-service-2 and validate Deferred policy allows one final response to be served by web-service-1
                idle_connection_test.go:612: [10.131.224.239:43382 -> 3.128.128.128:80] Req: URL=http://3.128.128.128, Host=test-idle-connection-close-deferred-t2mgq.apps.ci-op-q14lpbwi-9e7c5.origin-ci-int-aws.dev.rhcloud.com
                idle_connection_test.go:612: [10.131.224.239:43382 <- 3.128.128.128:80] Res: Status=503, Headers=map[Cache-Control:[private, max-age=0, no-cache, no-store] Content-Type:[text/html] Pragma:[no-cache]]
                idle_connection_test.go:598: step 2: unexpected response: got "", want "web-service-1"
                idle_connection_test.go:397: deleted ingresscontroller idle-connection-close-deferred-t2mgq
```
