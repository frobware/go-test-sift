`go-test-sift` is a tool for parsing and persisting Go unit test logs.
It simplifies the identification of failures, debugging of flaky
tests, and management of test output by serialising and grouping
output from parallel tests.

[ I primarily created this to help me triage e2e test failures from
various OpenShift repositories. ]

## Features

- **Serialising Parallel Unit Test Output**: Logically group test output from parallel tests.
- **Filter Output**: Focus on specific tests using regular expressions.
- **Summarise Failures**: Quickly identify which tests failed.
- **Serialise Logs**: Clean up interleaved output from parallel tests.
- **Save Logs**: Write individual test outputs to organised directories.

## Install

```
go install github.com/frobware/go-sift-tool@latest
```

## Usage

Run `go-test-sift` with your Go test log file or a URL as input:

```sh
$ go-test-sift [options] <GO-TEST-LOG-FILE | URL>
```

### Examples

Summarise test failures:

```sh
$ go-test-sift https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/pr-logs/pull/openshift_cluster-ingress-operator/1182/pull-ci-openshift-cluster-ingress-operator-master-e2e-aws-operator-techpreview/1881365030088216576/artifacts/e2e-aws-operator-techpreview/test/build-log.txt
--- FAIL: TestAll (3264.80s)
    --- FAIL: TestAll/parallel (96.31s)
        --- FAIL: TestAll/parallel/Test_IdleConnectionTerminationPolicy (0.00s)
            --- FAIL: TestAll/parallel/Test_IdleConnectionTerminationPolicy/Deferred (192.40s)
```

Show test output for test failures

```sh
$ go-test-sift -L https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/pr-logs/pull/openshift_cluster-ingress-operator/1182/pull-ci-openshift-cluster-ingress-operator-master-e2e-aws-operator-techpreview/1881365030088216576/artifacts/e2e-aws-operator-techpreview/test/build-log.txt
Failed Tests:
--- FAIL: TestAll (3264.80s)
    --- FAIL: TestAll/parallel (96.31s)
        --- FAIL: TestAll/parallel/Test_IdleConnectionTerminationPolicy (0.00s)
                operator_test.go:4210: pod idle-connection-close-immediate-hcm75/web-service-1 not ready
                idle_connection_test.go:482: [10.128.38.217:40602 -> 34.193.221.214:80] Req: URL=http://34.193.221.214, Host=test-idle-connection-close-immediate-hcm75.apps.ci-op-5mv1ytrz-9e7c5.origin-ci-int-aws.dev.rhcloud.com
                idle_connection_test.go:482: [10.128.38.217:40602 <- 34.193.221.214:80] Res: Status=200, Headers=map[Content-Length:[8] Content-Type:[text/plain; charset=utf-8] Date:[Mon, 20 Jan 2025 17:12:20 GMT] Set-Cookie:[1bcef74d604e5258d069e677f7d76762=5d96f964471d7a777df8f4c199e7958e; path=/; HttpOnly] X-Pod-Name:[web-service-2] X-Pod-Namespace:[unknown-namespace]]
                idle_connection_test.go:509: [10.128.38.217:34842 -> 100.28.61.223:80] Req: URL=http://100.28.61.223, Host=test-idle-connection-close-deferred-59n79.apps.ci-op-5mv1ytrz-9e7c5.origin-ci-int-aws.dev.rhcloud.com
                idle_connection_test.go:509: [10.128.38.217:34842 <- 100.28.61.223:80] Res: Status=503, Headers=map[Cache-Control:[private, max-age=0, no-cache, no-store] Content-Type:[text/html] Pragma:[no-cache]]
            --- FAIL: TestAll/parallel/Test_IdleConnectionTerminationPolicy/Deferred (192.40s)
                    idle_connection_test.go:540: Creating namespace "idle-connection-close-deferred-59n79"...
                    idle_connection_test.go:540: Waiting for ServiceAccount idle-connection-close-deferred-59n79/default to be provisioned...
                    idle_connection_test.go:540: Waiting for RoleBinding idle-connection-close-deferred-59n79/system:image-pullers to be created...
                    idle_connection_test.go:558: Creating IngressController openshift-ingress-operator/idle-connection-close-deferred-59n79...
                    util_test.go:694: waiting for loadbalancer domain a3d78b62aed85486aad9419d915f9cbd-2078764696.us-east-1.elb.amazonaws.com to resolve...
                    operator_test.go:4210: pod idle-connection-close-immediate-hcm75/web-service-1 not ready
                    operator_test.go:4210: pod idle-connection-close-immediate-hcm75/web-service-1 not ready
                    operator_test.go:4210: pod idle-connection-close-immediate-hcm75/web-service-2 not ready
                    operator_test.go:4210: pod idle-connection-close-immediate-hcm75/web-service-2 not ready
...
```

Use `-t` to filter by test name:

```sh
% ./go-test-sift -s -t TestRouteHardStopAfterTestOneDayDuration https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/pr-logs/pull/openshift_cluster-ingress-operator/1182/pull-ci-openshift-cluster-ingress-operator-master-e2e-aws-operator-techpreview/1881365030088216576/artifacts/e2e-aws-operator-techpreview/test/build-log.txt
        --- PASS: TestAll/serial/TestRouteHardStopAfterTestOneDayDuration (2.10s)
            2025-01-20T17:35:28.408Z	ERROR	operator.ingress_controller	ingress/deployment.go:215	invalid HAProxy time value	{"annotation": "ingress.operator.openshift.io/hard-stop-after", "value": "mañana", "error": "time: invalid duration \"ma\\xc3\\xb1ana\""}
```

Serialise all test output from parallel unit tests:

```sh
$ go-test-sift -s https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/pr-logs/pull/openshift_cluster-ingress-operator/1182/pull-ci-openshift-cluster-ingress-operator-master-e2e-aws-operator-techpreview/1881365030088216576/artifacts/e2e-aws-operator-techpreview/test/build-log.txt
<...outputs all tests with their respective outputs grouped together...>
```

Organise each test's output into its own directory:

```sh
$ ./go-test-sift -w -F https://gcsweb-ci.apps.ci.l2s4.p1.openshiftapps.com/gcs/test-platform-results/pr-logs/pull/openshift_cluster-ingress-operator/1182/pull-ci-openshift-cluster-ingress-operator-master-e2e-aws-operator-techpreview/1881365030088216576/artifacts/e2e-aws-operator-techpreview/test/build-log.txt
# tree output truncated
TestAll
├── parallel
│   ├── TestAWSEIPAllocationsForNLB
│   ├── TestAWSELBConnectionIdleTimeout
│   ├── TestAWSLBSubnets
│   ├── TestUnsupportedConfigOverride
│   ├── TestUserDefinedIngressController
│   └── Test_IdleConnectionTerminationPolicy
└── serial
    ├── TestAWSLBTypeDefaulting
    ├── TestAWSResourceTagsChanged
    ├── TestRouterCompressionOperation
    └── TestUpdateDefaultIngressControllerSecret
105 directories
```
