# Building the NVIDIA-exact ARM `virt-launcher`

This branch (`pvm-1393-aie-exact-image`, based on `release-1.9-aie-nv`) builds an **aarch64**
`virt-launcher` image with the exact userspace versions from the NVIDIA Grace/Vera guide and the
PVM-1393 launcher patches.

It is not a general KubeVirt build guide. It documents the image used for the VMO / NVIDIA
certification work.

## What the image contains

| Component | Version |
|---|---|
| KubeVirt base | `release-1.9-aie-nv` (KubeVirt 1.9.x) |
| QEMU | `qemu-kvm-11.0.0-1.el10nv` |
| libvirt | `libvirt-12.4.0-1.el10nv` |
| QEMU/libvirt source | NVIDIA `qemu 1:11.0.0+nvidia12-1`, `libvirt 12.4.0+nvidia-stable2-1` |
| Launcher patches | PCI hostdev `mode=subsystem`; `<qemu:override>` preservation |
| Extra rootfs package | `ipxe-roms-qemu` (needed for `efi-virtio.rom`) |

The image is **aarch64 only**. It does not contain a kernel. The host and guest kernels are separate.

## Prerequisites

- An aarch64 build host, or the KubeVirt CentOS Stream 10 builder image:
  `quay.io/kubevirt/builder-cs10:2602251001-25ce1ccb15`
- Bazel 6.5.x (present in the builder image)
- Network access for Bazel external dependencies
- The 13 exact aarch64 RPMs listed below, or a repo URL containing them
- Optional: `crane` for pushing the OCI image without a Docker daemon

## Required RPMs

The branch's `WORKSPACE` expects these aarch64 RPMs at `/rpms`:

```
libvirt-client-12.4.0-1.el10nv.aarch64.rpm
libvirt-daemon-common-12.4.0-1.el10nv.aarch64.rpm
libvirt-daemon-driver-qemu-12.4.0-1.el10nv.aarch64.rpm
libvirt-daemon-log-12.4.0-1.el10nv.aarch64.rpm
libvirt-libs-12.4.0-1.el10nv.aarch64.rpm
libvirt-devel-12.4.0-1.el10nv.aarch64.rpm
qemu-img-11.0.0-1.el10nv.aarch64.rpm
qemu-kvm-common-11.0.0-1.el10nv.aarch64.rpm
qemu-kvm-core-11.0.0-1.el10nv.aarch64.rpm
qemu-kvm-device-display-virtio-gpu-11.0.0-1.el10nv.aarch64.rpm
qemu-kvm-device-display-virtio-gpu-pci-11.0.0-1.el10nv.aarch64.rpm
qemu-kvm-device-usb-host-11.0.0-1.el10nv.aarch64.rpm
qemu-kvm-device-usb-redirect-11.0.0-1.el10nv.aarch64.rpm
```

`WORKSPACE` currently points at `file:///rpms/<name>.rpm`. Either:

- copy the RPMs to `/rpms` on the build host, or
- change those URLs to your team's durable RPM repository.

Do not merge the branch with `file:///rpms` URLs unless the CI/build environment provides that path.

## Rebuilding the exact RPMs from NVIDIA source

The exact Ubuntu source packages are published by NVIDIA:

- QEMU: `qemu 1:11.0.0+nvidia12-1`
  - `qemu_11.0.0+nvidia12-1.dsc`
  - `qemu_11.0.0+nvidia12.orig.tar.gz`
  - `qemu_11.0.0+nvidia12-1.debian.tar.xz`
- libvirt: `libvirt 12.4.0+nvidia-stable2-1`
  - `libvirt_12.4.0+nvidia-stable2-1.dsc`
  - `libvirt_12.4.0+nvidia-stable2.orig.tar.gz`
  - `libvirt_12.4.0+nvidia-stable2-1.debian.tar.xz`

They are available from the Launchpad PPA:

```
https://launchpad.net/~canonical-nvidia/+archive/ubuntu/nvidia-virtualization
```

Extract them with:

```shell
dpkg-source -x qemu_11.0.0+nvidia12-1.dsc qemu-11.0.0
dpkg-source -x libvirt_12.4.0+nvidia-stable2-1.dsc libvirt-12.4.0
```

The extracted trees already contain NVIDIA's patches. We built aarch64 CentOS Stream 10 RPMs from
those trees with the AIE dist-git specs:

- `qemu-kvm.spec` from `c10s-aie-nv` (version changed to `11.0.0`, dist `.el10nv`)
- `libvirt.spec` from `c10s-aie-nv` (version changed to `12.4.0`, dist `.el10nv`)

Important build notes:

- Do **not** re-apply the dist-git patch series; the PPA source tree already contains the NVIDIA
  changes.
- Disable debuginfo packages (`%global debug_package %{nil}`) if you do not need them.
- QEMU needs `--with-devices-aarch64=default` when building from this source without the RHEL
  device-list patch. Disable `--enable-werror` if your compiler reports unused-variable errors.
- The libvirt soname is `libvirt.so.0.12004.0` for 12.4.0. The branch updates the `rpm/BUILD.bazel`
  references from `12005` to `12004`.
- The QEMU RPMs must include `ipxe-roms-qemu`; without it the launcher fails with
  `failed to find romfile "efi-virtio.rom"`.

## Build on an aarch64 host

```shell
cd ~/Desktop/spectro/kubevirt
git switch pvm-1393-aie-exact-image

# Provide the exact RPMs at /rpms (or edit WORKSPACE to your repo).
sudo mkdir -p /rpms
sudo cp /path/to/rpms/*.aarch64.rpm /rpms/

bazel build \
  --config=aarch64 \
  --config=cs10 \
  --jobs=32 \
  //cmd/virt-launcher:virt-launcher-image
```

Output:

```
bazel-bin/cmd/virt-launcher/virt-launcher-image
```

This is an OCI image layout (`index.json`, `oci-layout`, `blobs/`).

## Build in the KubeVirt builder container

If the source tree was copied without `.git`, Bazel's workspace-status script needs these env vars:

```shell
docker run --rm -it \
  --platform linux/arm64 \
  -v "$PWD":/workspace/kubevirt \
  -v /rpms:/rpms \
  -w /workspace/kubevirt \
  quay.io/kubevirt/builder-cs10:2602251001-25ce1ccb15 \
  bash -lc '
    export KUBEVIRT_GIT_VERSION=v1.9.1
    export KUBEVIRT_GIT_COMMIT=unknown
    export KUBEVIRT_GIT_TREE_STATE=clean
    export KUBEVIRT_BOOTSTRAPPING=true
    bazel build --config=aarch64 --config=cs10 --jobs=32 \
      //cmd/virt-launcher:virt-launcher-image
  '
```

## Push the image

Using `crane` directly from the OCI layout:

```shell
crane push bazel-bin/cmd/virt-launcher/virt-launcher-image \
  <registry>/<project>/virt-launcher:pvm-1393-aie-exact-<date>
```

If your registry requires a Docker credential helper, you can push to a temporary local registry and
then use Docker:

```shell
docker run -d -p 5000:5000 --name tmp-registry registry:2
crane push --insecure bazel-bin/cmd/virt-launcher/virt-launcher-image \
  127.0.0.1:5000/virt-launcher:local
docker pull 127.0.0.1:5000/virt-launcher:local
docker tag 127.0.0.1:5000/virt-launcher:local \
  <registry>/<project>/virt-launcher:pvm-1393-aie-exact-<date>
docker push <registry>/<project>/virt-launcher:pvm-1393-aie-exact-<date>
```

## Test on the ARM cluster

After pushing, point KubeVirt at the image. For a lab test, patch the `KubeVirt` CR:

```yaml
spec:
  customizeComponents:
    patches:
    - type: json
      resourceType: Deployment
      resourceName: virt-controller
      patch: '[{"op":"replace","path":"/spec/template/spec/containers/0/args/1","value":"<image>"}]'
```

If the registry is private, add its pull secret to the service account used by the VMI
(`kubectl patch sa default -n <ns> -p '{"imagePullSecrets":[{"name":"regcred"}]}'`).

Start the VM and verify:

```shell
# inside the launcher pod
virsh --version
/usr/libexec/qemu-kvm --version

# inside the guest (virtctl ssh or qemu-guest-agent)
nvidia-smi -L
nvidia-smi --query-gpu=name,driver_version,memory.total --format=csv,noheader
nvidia-smi -q | grep -A2 'BAR1 Memory Usage'
nvidia-smi topo -m
```

Expected for the T4G validation:

- libvirt `12.4.0`
- QEMU `11.0.0`
- BAR1 `256 MiB`
- `nvidia-smi topo -m` shows `PHB`
- domain XML contains `pcihole64` and `<qemu:override>` with `pref64-reserve`

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `failed to find romfile "efi-virtio.rom"` | `ipxe-roms-qemu` missing from the launcher rootfs; add it to `launcherbase_aarch64_cs10`, and keep the `/usr/share/qemu` symlinks in `cmd/virt-launcher/BUILD.bazel`. |
| `libvirt.so.0.12005.0` not found | Your libvirt is 12.4.0, whose soname is `12004`; update `rpm/BUILD.bazel` references. |
| `KUBEVIRT_GIT_VERSION: unbound variable` | The build tree has no `.git`; set the version env vars shown above. |
| `file:///rpms/...` cannot be downloaded | Copy the RPMs to `/rpms` or change `WORKSPACE` to your artifact repository. |
| Bazel tries to fetch debuginfo RPMs | Do not set `KUBEVIRT_AIE_NV_INCLUDE_DEBUGINFO=true` unless you built the corresponding debuginfo packages. |
| VM starts but no GPU | Check `permittedHostDevices`, the VFIO binding, and the exact resource names; this guide assumes the GPU passthrough setup is already working. |

## Scope

This build proves the ARM launcher userspace and the QEMU/libvirt versions. It does **not** by itself
satisfy the Grace/GB200 platform requirements: `nvgrace-gpu-vfio-pci`, SMMUv3, IOMMUFD, ACPI Generic
Initiator NUMA nodes, 512M hugepages, GICv4 and the 64k kernel are validated separately on the GB200
lab.
