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

Publication and anonymous pull must be confirmed before using these tags.
The workflow summary records the immutable registry digests for each build.
The embedded version string remains CoreDNS 1.14.7; use the commit label and
image digest to identify these development builds.

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

The local experiment reproduces a first TLS-handshake stall when ICMP PMTU
notifications reach the client but not the upstream. The before image's source
fails the first query after about30s; the after image's source reconnects and
succeeds after about2.5s. A fresh SYN advertises a smaller MSS after the client
learns the PMTU. This explains one failure mechanism, not the exact cause of
the reporter's Talos environment or every form of PMTU black hole.
