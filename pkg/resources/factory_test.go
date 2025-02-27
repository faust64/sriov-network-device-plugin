package resources

import (
	"reflect"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"

	"github.com/intel/sriov-network-device-plugin/pkg/types"
	"github.com/intel/sriov-network-device-plugin/pkg/types/mocks"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("Factory", func() {
	Describe("getting factory instance", func() {
		Context("always", func() {
			It("should return the same instance", func() {
				f0 := NewResourceFactory("fake", "fake", true)
				Expect(f0).NotTo(BeNil())
				f1 := NewResourceFactory("fake", "fake", true)
				Expect(f1).To(Equal(f0))
			})
		})
	})
	DescribeTable("getting info provider",
		func(name string, expected reflect.Type) {
			f := NewResourceFactory("fake", "fake", true)
			p := f.GetInfoProvider(name)
			Expect(reflect.TypeOf(p)).To(Equal(expected))
		},
		Entry("vfio-pci", "vfio-pci", reflect.TypeOf(&vfioResourcePool{})),
		Entry("uio", "uio", reflect.TypeOf(&uioResourcePool{})),
		Entry("any other value", "netdevice", reflect.TypeOf(&netDevicePool{})),
	)
	Describe("getting resource pool", func() {
		Context("with all types of selectors used", func() {
			var (
				rp   types.ResourcePool
				err  error
				devs []types.PciNetDevice
			)
			BeforeEach(func() {
				f := NewResourceFactory("fake", "fake", true)

				devs = make([]types.PciNetDevice, 4)
				vendors := []string{"8086", "8086", "8086", "1234"}
				codes := []string{"1111", "1111", "1234", "4321"}
				drivers := []string{"vfio-pci", "i40evf", "igb_uio", "igb_uio"}
				pfNames := []string{"enp2s0f2", "ens0", "eth0", "net2"}
				for i := range devs {
					d := &mocks.PciNetDevice{}
					d.On("GetVendor").Return(vendors[i]).
						On("GetDeviceCode").Return(codes[i]).
						On("GetDriver").Return(drivers[i]).
						On("GetPFName").Return(pfNames[i]).
						On("GetPciAddr").Return("fake").
						On("GetAPIDevice").Return(&pluginapi.Device{})
					devs[i] = d
				}

				c := types.ResourceConfig{
					ResourceName: "fake",
					Selectors: struct {
						Vendors []string `json:"vendors,omitempty"`
						Devices []string `json:"devices,omitempty"`
						Drivers []string `json:"drivers,omitempty"`
						PfNames []string `json:"pfNames,omitempty"`
					}{[]string{"8086"}, []string{"1111"}, []string{"vfio-pci"}, []string{"enp2s0f2"}},
				}

				rp, err = f.GetResourcePool(&c, devs)
			})
			It("should return valid resource pool", func() {
				Expect(rp).NotTo(BeNil())
				Expect(rp.(*resourcePool).devices).To(HaveLen(1))
				Expect(rp.(*resourcePool).devices).To(HaveKey("fake"))
				Expect(rp.(*resourcePool).devicePool).To(HaveKeyWithValue("fake", devs[0]))
			})
			It("should not fail", func() {
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})
	Describe("getting rdma spec", func() {
		Context("check c rdma spec", func() {
			f := NewResourceFactory("fake", "fake", true)
			rs := f.GetRdmaSpec("0000:00:00.1")
			isRdma := rs.IsRdma()
			deviceSpec := rs.GetRdmaDeviceSpec()
			It("shoud return valid rdma spec", func() {
				Expect(isRdma).ToNot(BeTrue())
				Expect(deviceSpec).To(HaveLen(0))
			})
		})
	})
})
