/*
 * This file is part of the KubeVirt project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright The KubeVirt Authors.
 *
 */

package api

// Hand-written deepcopy helpers for the <qemu:override> types added to
// DomainSpec. Running deepcopy-gen would produce the same methods.

func (in *QEMUOverride) DeepCopyInto(out *QEMUOverride) {
	*out = *in
	if in.Devices != nil {
		in, out := &in.Devices, &out.Devices
		*out = make([]QEMUOverrideDevice, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func (in *QEMUOverride) DeepCopy() *QEMUOverride {
	if in == nil {
		return nil
	}
	out := new(QEMUOverride)
	in.DeepCopyInto(out)
	return out
}

func (in *QEMUOverrideDevice) DeepCopyInto(out *QEMUOverrideDevice) {
	*out = *in
	in.Frontend.DeepCopyInto(&out.Frontend)
}

func (in *QEMUOverrideDevice) DeepCopy() *QEMUOverrideDevice {
	if in == nil {
		return nil
	}
	out := new(QEMUOverrideDevice)
	in.DeepCopyInto(out)
	return out
}

func (in *QEMUOverrideFrontend) DeepCopyInto(out *QEMUOverrideFrontend) {
	*out = *in
	if in.Properties != nil {
		in, out := &in.Properties, &out.Properties
		*out = make([]QEMUOverrideProperty, len(*in))
		copy(*out, *in)
	}
}

func (in *QEMUOverrideFrontend) DeepCopy() *QEMUOverrideFrontend {
	if in == nil {
		return nil
	}
	out := new(QEMUOverrideFrontend)
	in.DeepCopyInto(out)
	return out
}

func (in *QEMUOverrideProperty) DeepCopyInto(out *QEMUOverrideProperty) {
	*out = *in
}

func (in *QEMUOverrideProperty) DeepCopy() *QEMUOverrideProperty {
	if in == nil {
		return nil
	}
	out := new(QEMUOverrideProperty)
	in.DeepCopyInto(out)
	return out
}
