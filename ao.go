package ao

// #include <ao/ao.h>
import "C"
import (
	"os"
	"unsafe"
)

const (
	FormatLittle = C.AO_FMT_LITTLE
	FormatBig    = C.AO_FMT_BIG
	FormatNative = C.AO_FMT_NATIVE
)

type Option struct {
	Key   string
	Value string
}

type Device struct {
	ao *C.ao_device
}

type SampleFormat struct {
	Bits       int    // bits per sample
	Rate       int    // samples per second (in a single channel)
	Channels   int    // number of audio channels
	ByteFormat int    // byte ordering in sample
	Matrix     string // channel input matrix
}

type Info struct {
	Live                bool
	Name                string
	ShortName           string
	Comment             string
	PreferredByteFormat int
	Priority            int
	Options             []string
}

type Errno int64

var (
	ENODRIVER   os.Error = Errno(C.AO_ENODRIVER)
	ENOTLIVE    os.Error = Errno(C.AO_ENOTLIVE)
	ENOTFILE    os.Error = Errno(C.AO_ENOTFILE)
	EBADOPTION  os.Error = Errno(C.AO_EBADOPTION)
	EOPENDEVICE os.Error = Errno(C.AO_EOPENDEVICE)
	EFAIL       os.Error = Errno(C.AO_EFAIL)
)

func (e Errno) String() string {
	switch int64(e) {
	case C.AO_ENODRIVER:
		return "no driver corresponds to driver id"
	case C.AO_ENOTLIVE:
		return "this driver is not a live output device"
	case C.AO_ENOTFILE:
		return "this driver is not a file output driver"
	case C.AO_EBADOPTION:
		return "a valid option key has an invalid value"
	case C.AO_EOPENDEVICE:
		return "cannot open the device"
	}
	return "libao failure"
}

func Initialize() {
	C.ao_initialize()
}

func Shutdown() {
	C.ao_shutdown()
}

func DefaultDriverID() int {
	return int(C.ao_default_driver_id())
}

func DriverID(shortName string) int {
	short_name := C.CString(shortName)
	result := C.ao_driver_id(short_name)
	C.free(unsafe.Pointer(short_name))
	return int(result)
}

func fillInfo(info *C.ao_info) *Info {
	i := Info{
		Live:                info._type == C.AO_TYPE_LIVE,
		Name:                C.GoString(info.name),
		ShortName:           C.GoString(info.short_name),
		Comment:             C.GoString(info.comment),
		PreferredByteFormat: int(info.preferred_byte_format),
		Priority:            int(info.priority),
		Options:             make([]string, info.option_count),
	}
	opts := uintptr(unsafe.Pointer(info.options))
	for x := 0; x < int(info.option_count); x++ {
		i.Options[x] = C.GoString(*(**C.char)(unsafe.Pointer(opts)))
		opts += uintptr(unsafe.Sizeof(info.options))
	}
	return &i
}

func DriverInfo(id int) (*Info, os.Error) {
	info := C.ao_driver_info(C.int(id))
	if info == nil {
		return nil, ENODRIVER
	}
	return fillInfo(info), nil
}

func DriverInfoList() []Info {
	var count C.int
	infos := C.ao_driver_info_list(&count)
	p := uintptr(unsafe.Pointer(infos))
	result := make([]Info, count)
	for i, _ := range result {
		result[i] = *fillInfo(*(**C.ao_info)(unsafe.Pointer(p)))
		p += uintptr(unsafe.Sizeof(infos))
	}
	return result
}

func appendOption(copt **C.ao_option, opt *Option) int {
	key := C.CString(opt.Key)
	value := C.CString(opt.Value)
	result := C.ao_append_option(copt, key, value)
	C.free(unsafe.Pointer(key))
	C.free(unsafe.Pointer(value))
	return int(result)
}

func OpenLive(driverID int, format *SampleFormat, options ...Option) (d *Device, e os.Error) {
	var opts *C.ao_option
	for _, o := range options {
		if appendOption(&opts, &o) == 0 {
			C.ao_free_options(opts)
			return nil, os.ENOMEM
		}
	}
	fmt := C.ao_sample_format{
		bits:        C.int(format.Bits),
		rate:        C.int(format.Rate),
		channels:    C.int(format.Channels),
		byte_format: C.int(format.ByteFormat),
		matrix:      C.CString(format.Matrix),
	}
	d = &Device{}
	d.ao, e = C.ao_open_live(C.int(driverID), &fmt, opts)
	if d.ao != nil {
		e = nil
	}
	C.free(unsafe.Pointer(fmt.matrix))
	C.ao_free_options(opts)
	return
}

func OpenFile(driverID int, path string, overwrite bool, format *SampleFormat, options ...Option) (d *Device, e os.Error) {
	var opts *C.ao_option
	for _, o := range options {
		if appendOption(&opts, &o) == 0 {
			C.ao_free_options(opts)
			return nil, os.ENOMEM
		}
	}
	filename := C.CString(path)
	overwrite_ := C.int(0)
	if overwrite {
		overwrite_ = 1
	}
	fmt := C.ao_sample_format{
		bits:        C.int(format.Bits),
		rate:        C.int(format.Rate),
		channels:    C.int(format.Channels),
		byte_format: C.int(format.ByteFormat),
		matrix:      C.CString(format.Matrix),
	}
	d = &Device{}
	d.ao, e = C.ao_open_file(C.int(driverID), filename, overwrite_, &fmt, opts)
	if d.ao != nil {
		e = nil
	}
	C.free(unsafe.Pointer(fmt.matrix))
	C.free(unsafe.Pointer(filename))
	C.ao_free_options(opts)
	return
}

func (device *Device) Play8(samples []byte) bool {
	raw := (*C.char)(unsafe.Pointer(&samples[0]))
	num := C.uint_32(len(samples))
	return C.ao_play(device.ao, raw, num) != 0
}

func (device *Device) Play16(samples []int16) bool {
	raw := (*C.char)(unsafe.Pointer(&samples[0]))
	num := C.uint_32(len(samples) * 2)
	return C.ao_play(device.ao, raw, num) != 0
}

func (device *Device) Close() bool {
	return C.ao_close(device.ao) != 0
}
