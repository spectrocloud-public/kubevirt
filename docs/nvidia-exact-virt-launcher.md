# Building the ARM `virt-launcher` with upstream AIE RPMs

This branch (`pvm-1393-aie-exact-image`, based on `release-1.9-aie-nv`) builds an **aarch64**
`virt-launcher` image for the VMO / NVIDIA certification work.

It uses the **existing upstream CentOS Stream AIE `el10nv` RPMs**, not locally rebuilt packages.
The selected versions are at or above the NVIDIA Grace/Vera guide minimums.

## Versions

| Component | Version | Note |
|---|---|---|
| KubeVirt base | `release-1.9-aie-nv` | KubeVirt 1.9.x |
| QEMU | `qemu-kvm-11.1.0-3.el10nv` | above guide's 11.0.0 |
| libvirt | `libvirt-12.5.0-1.el10nv` | above guide's 12.4.0 |
| libvirt soname | `libvirt.so.0.12005.0` | matches libvirt 12.5 |
| Launcher patches | PCI hostdev `mode=subsystem`; `<qemu:override>` preservation | |

The image is **aarch64 only**. It does not contain a kernel. The host and guest kernels are separate.

## Validated image

Built and tested on the ARM AWS cluster:

```text
us-east1-docker.pkg.dev/spectro-images/dev/vmo/ahmed/nvda-cert/virt-launcher:pvm-1393-aie-upstream-20261120
sha256:7dffb4cda18e4fce6160e2a8283812760ed0870b920ebbdb19294f6424f7840a
```

T4G validation: 2× T4G, driver 595.91.07, BAR1 256 MiB, `topo` PHB, analyzer 0 FAIL / 1 WARN
(host topology not supplied).

## RPM source

The RPMs are published by the CentOS Stream AIE SIG:

```text
https://odcs.stream.centos.org/stream-10/aie/nv/aarch64/
https://odcs.stream.centos.org/stream-10/aie/nv/x86_64/
```

`WORKSPACE` contains the exact `rpm()` entries and sha256 values. There is no `/rpms` staging step
and no local RPM build.

## Prerequisites

- An aarch64 build host, or the KubeVirt CentOS Stream 10 builder image:
  `quay.io/kubevirt/builder-cs10:2602251001-25ce1ccb15`
- Bazel 6.5.x (present in the builder image)
- Network access for Bazel external dependencies and the AIE RPMs
- Optional: `crane` for pushing the OCI image without a Docker daemon

## 1. Checkout the branch

```bash
git clone git@github.com:spectrocloud-public/kubevirt.git
cd kubevirt

git fetch origin release-1.9-aie-nv
git fetch origin pvm-1393-aie-exact-image
git switch pvm-1393-aie-exact-image
```

## 2. Refresh or inspect the RPM entries

The branch already contains the required `WORKSPACE` and `rpm/BUILD.bazel` entries. To regenerate them
from the AIE repo:

```bash
COMPOSE_BASE_URL=https://odcs.stream.centos.org/stream-10/aie/nv \
LIBVIRT_NV_VERSION=12.5.0 LIBVIRT_NV_RELEASE=1.el10nv \
QEMU_NV_VERSION=11.1.0 QEMU_NV_RELEASE=3.el10nv \
hack/sync-nv-rpms.sh
```

`hack/sync-nv-rpms.sh` needs `curl`, `rpm` and `sha256sum`. It downloads both `aarch64` and `x86_64`
RPMs, verifies them, and updates `WORKSPACE` and `rpm/BUILD.bazel`.

The packages used by the aarch64 launcher are:

```text
libvirt-client-12.5.0-1.el10nv.aarch64.rpm
libvirt-daemon-common-12.5.0-1.el10nv.aarch64.rpm
libvirt-daemon-driver-qemu-12.5.0-1.el10nv.aarch64.rpm
libvirt-daemon-log-12.5.0-1.el10nv.aarch64.rpm
libvirt-libs-12.5.0-1.el10nv.aarch64.rpm
libvirt-devel-12.5.0-1.el10nv.aarch64.rpm
qemu-img-11.1.0-3.el10nv.aarch64.rpm
qemu-kvm-common-11.1.0-3.el10nv.aarch64.rpm
qemu-kvm-core-11.1.0-3.el10nv.aarch64.rpm
qemu-kvm-device-display-virtio-gpu-11.1.0-3.el10nv.aarch64.rpm
qemu-kvm-device-display-virtio-gpu-pci-11.1.0-3.el10nv.aarch64.rpm
qemu-kvm-device-usb-host-11.1.0-3.el10nv.aarch64.rpm
qemu-kvm-device-usb-redirect-11.1.0-3.el10nv.aarch64.rpm
```

`ipxe-roms-qemu` is already in the aarch64 launcher rootfs. It provides `efi-virtio.rom`, which
KubeVirt's default virtio-net device needs.

## 3. Build on an aarch64 host

```bash
cd kubevirt

bazel build \
  --config=aarch64 \
  --config=cs10 \
  --jobs=32 \
  //cmd/virt-launcher:virt-launcher-image
```

Output:

```text
bazel-bin/cmd/virt-launcher/virt-launcher-image
```

This is an OCI image layout (`index.json`, `oci-layout`, `blobs/`).

## 4. Build with the KubeVirt builder container

```bash
docker run --rm -it \
  --platform linux/arm64 \
  -v "$PWD":/workspace/kubevirt \
  -w /workspace/kubevirt \
  quay.io/kubevirt/builder-cs10:2602251001-25ce1ccb15 \
  bash -lc '
    export KUBEVIRT_GIT_VERSION=v1.9.1
    export KUBEVIRT_GIT_COMMIT=unknown
    export KUBEVIRT_GIT_TREE_STATE=clean
    export KUBEVIRT_BOOTSTRAPPING=true

    bazel build \
      --config=aarch64 \
      --config=cs10 \
      --jobs=32 \
      //cmd/virt-launcher:virt-launcher-image
  '
```

The `KUBEVIRT_GIT_*` variables are only needed when the source tree has no `.git` directory.

For an x86 build host with cross-compilation, use the cross builder image and:

```bash
bazel build \
  --config=crossbuild-aarch64 \
  --config=cs10 \
  --jobs=32 \
  //cmd/virt-launcher:virt-launcher-image
```

## 5. Push the image

```bash
IMAGE=<registry>/<project>/virt-launcher:pvm-1393-aie-$(date +%Y%m%d)

crane push \
  bazel-bin/cmd/virt-launcher/virt-launcher-image \
  "$IMAGE"
```

Example registry:

```text
us-east1-docker.pkg.dev/spectro-images/dev/vmo/ahmed/nvda-cert/virt-launcher:pvm-1393-aie-20261120
```

If `crane` cannot use your registry credential helper, use a temporary local registry:

```bash
docker run -d -p 5000:5000 --name tmp-registry registry:2

crane push --insecure \
  bazel-bin/cmd/virt-launcher/virt-launcher-image \
  127.0.0.1:5000/virt-launcher:local

docker pull 127.0.0.1:5000/virt-launcher:local
docker tag 127.0.0.1:5000/virt-launcher:local "$IMAGE"
docker push "$IMAGE"
```

## 6. Point KubeVirt at the image

For a lab test, patch the `KubeVirt` CR:

```yaml
spec:
  customizeComponents:
    patches:
    - type: json
      resourceType: Deployment
      resourceName: virt-controller
      patch: '[{"op":"replace","path":"/spec/template/spec/containers/0/args/1","value":"<image>"}]'
```

If the registry is private, add its pull secret to the service account used by the VMI:

```bash
kubectl patch sa default -n default \
  -p '{"imagePullSecrets":[{"name":"regcred"}]}'
```

## 7. Verify the image

Inside the launcher pod:

```bash
virsh --version
/usr/libexec/qemu-kvm --version
```

Expected:

```text
12.5.0
QEMU emulator version 11.1.0 (qemu-kvm-11.1.0-3.el10nv)
```

Inside the guest:

```bash
nvidia-smi -L
nvidia-smi --query-gpu=name,driver_version,memory.total --format=csv,noheader
nvidia-smi -q | grep -A2 "BAR1 Memory Usage"
nvidia-smi topo -m
```

Expected for the T4G validation:

```text
2x NVIDIA T4G
driver 595.91.07
BAR1 256 MiB
topo: PHB
```

Check the domain XML:

```bash
virsh dumpxml default_<vm-name> | grep -A8 'qemu:override'
```

Expected: `pref64-reserve` entries.

## 8. Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `failed to find romfile "efi-virtio.rom"` | `ipxe-roms-qemu` missing from the launcher rootfs; keep it in `launcherbase_aarch64_cs10` and keep the `/usr/share/qemu` symlinks in `cmd/virt-launcher/BUILD.bazel`. |
| `libvirt.so.0.12005.0` not found | `rpm/BUILD.bazel` must reference the libvirt 12.5 soname (`12005`). |
| `KUBEVIRT_GIT_VERSION: unbound variable` | The build tree has no `.git`; set the version env vars shown above. |
| AIE RPM download fails | Check `https://odcs.stream.centos.org/stream-10/aie/nv/<arch>/Packages/` and rerun `hack/sync-nv-rpms.sh`. |
| Bazel tries to fetch debuginfo RPMs | Do not set `KUBEVIRT_AIE_NV_INCLUDE_DEBUGINFO=true` unless you need the debuginfo packages. |
| VM starts but no GPU | Check `permittedHostDevices`, the VFIO binding, and the exact resource names; this guide assumes GPU passthrough is already working. |

## 9. Scope

This image provides the ARM launcher userspace and QEMU/libvirt versions at or above the Grace/Vera
guide minimums. It does **not** by itself satisfy the GB200 platform requirements:

- `nvgrace-gpu-vfio-pci`
- SMMUv3
- IOMMUFD
- ACPI Generic Initiator NUMA nodes
- 512M hugepages
- GICv4 and firmware 1.3+ with CANWBS
- 64k kernel

Those are validated on the GB200 lab separately.
