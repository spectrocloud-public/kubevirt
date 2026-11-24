package translate

import (
	"encoding/xml"
	"strings"
	"testing"

	libvirtxml "libvirt.org/go/libvirtxml"
)

func TestQEMUOverrideRoundTrip(t *testing.T) {
	in := `<domain type="kvm">
  <name>t</name>
  <memory unit="b">1073741824</memory>
  <os><type arch="aarch64">hvm</type></os>
  <devices>
    <controller type="pci" index="0" model="pcie-root"/>
    <controller type="pci" index="1" model="pcie-root-port"><alias name="pci.1"/></controller>
  </devices>
  <override xmlns="http://libvirt.org/schemas/domain/qemu/1.0">
    <device alias="pci.1">
      <frontend>
        <property name="pref64-reserve" type="unsigned" value="21474836480"/>
      </frontend>
    </device>
  </override>
</domain>`
	d := &libvirtxml.Domain{}
	if err := d.Unmarshal(in); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if d.QEMUOverride == nil {
		t.Fatalf("libvirtxml did not parse qemu:override")
	}
	spec, err := FromLibvirtDomain(d)
	if err != nil {
		t.Fatalf("FromLibvirtDomain: %v", err)
	}
	if spec.QEMUOverride == nil || len(spec.QEMUOverride.Devices) != 1 {
		t.Fatalf("override lost in DomainSpec: %+v", spec.QEMUOverride)
	}
	out, err := xml.Marshal(spec)
	if err != nil {
		t.Fatalf("marshal spec: %v", err)
	}
	s := string(out)
	for _, want := range []string{"qemu:override", "xmlns:qemu", `alias="pci.1"`, `name="pref64-reserve"`, `value="21474836480"`} {
		if !strings.Contains(s, want) {
			t.Fatalf("marshaled domain missing %q:\n%s", want, s)
		}
	}
	back, err := ToLibvirtDomain(spec)
	if err != nil {
		t.Fatalf("ToLibvirtDomain: %v", err)
	}
	if back.QEMUOverride == nil {
		t.Fatalf("override lost going back to libvirtxml")
	}
}
