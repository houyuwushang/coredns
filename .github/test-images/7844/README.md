# Issue 7844 Comparison Images

This personal-fork branch builds unofficial linux/amd64 images for testing
[CoreDNS issue 7844](https://github.com/coredns/coredns/issues/7844) and
[PR 8543](https://github.com/coredns/coredns/pull/8543). It is not an official
CoreDNS release and is not intended for the production DNS Deployment.

The workflow checks out two fixed upstream commits separately from this
test harness, uses Go 1.27.0, checks binary hashes against the executables
used in an isolated Linux macvlan/PMTU experiment, and packages them with the
upstream Dockerfile. It also runs the related race tests and non-root UDP/TCP
container checks before publishing.

| Image Tag | Source Commit |
| --- | --- |
| `ghcr.io/houyuwushang/coredns-7844:before-19adcd8b9-amd64` | `19adcd8b960ebd349e4ccbdc04b1a12d77a7e0c0` |
| `ghcr.io/houyuwushang/coredns-7844:after-f59ebacbc-amd64` | `f59ebacbc7d95b2f9c48220371fa8f5607c8d9d8` |

Both images were published on 2026-09-12 by
[this workflow run](https://github.com/houyuwushang/coredns/actions/runs/34675692117).
Complete anonymous image downloads succeeded, and the extracted `/coredns`
executables match the binary hashes used in the Linux experiment. Both builds
passed the proxy/forward race tests and three UDP plus three TCP container
queries as user `65532:65532`.

Use these immutable references in the test Pod's `image` field:

Before PR 8543:

```text
ghcr.io/houyuwushang/coredns-7844@sha256:da5037df412f64b9fcb4b6f091a79a9ea3ef9bf18b6ab40aee20af7483e46c96
```

With PR 8543:

```text
ghcr.io/houyuwushang/coredns-7844@sha256:dc9e215d5386ca43a606d5749bf5154e3b7a3891efc594c8a60fea14c9ac7e8f
```

The embedded version string remains CoreDNS 1.14.7; use the commit label and
image digest to identify these development builds. Neither is a rebuild of
the originally reported CoreDNS 1.12.3 release.

Use a new, isolated test Pod for each image, with the same macvlan and VPN
egress settings as the affected deployment. Do not change a shared production
network attachment or use labels selected by the production DNS Service.
Avoid ping/OpenSSL queries to the upstream before the first DNS request:
these can prewarm the route PMTU cache and hide the failure.

Compare the first DNS query and an immediate follow-up with a 35-second query
timeout and one client attempt. The CoreDNS cache plugin can answer a repeated
query without forwarding; disable it only in the isolated test Corefile when
testing upstream recovery. Recreate the Pod rather than merely restarting the
CoreDNS process between image comparisons.

From a client permitted by the test Corefile's ACL, set `TEST_DNS_IP` to the
isolated test Pod's reachable address and run:

```sh
dig @"$TEST_DNS_IP" google.com A +time=35 +tries=1 +noall +comments +answer +stats
dig @"$TEST_DNS_IP" google.com A +time=35 +tries=1 +noall +comments +answer +stats
```

Keep both Pods on the same node/path when possible. Share each query's status
and latency, together with the corresponding CoreDNS errors. These Pod-level
instructions have not been run on the reporter's Talos cluster; the completed
network reproduction used isolated Linux namespaces.

The local experiment reproduces a first TLS-handshake stall when ICMP PMTU
notifications reach the client but not the upstream. The before image's source
fails the first query after about 30s; the after image's source reconnects and
succeeds after about 2.5s. A fresh SYN advertises a smaller MSS after the client
learns the PMTU. This explains one failure mechanism, not the exact cause of
the reporter's Talos environment or every form of PMTU black hole.
