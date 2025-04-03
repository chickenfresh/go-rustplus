// Copyright 2014 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.
//
// Request and reply to the "checkin server" devices poll every few hours.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v5.29.3
// source: checkin.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// A concrete name/value pair sent to the device's Gservices database.
type GservicesSetting struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          []byte                 `protobuf:"bytes,1,req,name=name" json:"name,omitempty"`
	Value         []byte                 `protobuf:"bytes,2,req,name=value" json:"value,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GservicesSetting) Reset() {
	*x = GservicesSetting{}
	mi := &file_checkin_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GservicesSetting) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GservicesSetting) ProtoMessage() {}

func (x *GservicesSetting) ProtoReflect() protoreflect.Message {
	mi := &file_checkin_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GservicesSetting.ProtoReflect.Descriptor instead.
func (*GservicesSetting) Descriptor() ([]byte, []int) {
	return file_checkin_proto_rawDescGZIP(), []int{0}
}

func (x *GservicesSetting) GetName() []byte {
	if x != nil {
		return x.Name
	}
	return nil
}

func (x *GservicesSetting) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

// Devices send this every few hours to tell us how they're doing.
type AndroidCheckinRequest struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// IMEI (used by GSM phones) is sent and stored as 15 decimal
	// digits; the 15th is a check digit.
	Imei *string `protobuf:"bytes,1,opt,name=imei" json:"imei,omitempty"` // IMEI, reported but not logged.
	// MEID (used by CDMA phones) is sent and stored as 14 hexadecimal
	// digits (no check digit).
	Meid *string `protobuf:"bytes,10,opt,name=meid" json:"meid,omitempty"` // MEID, reported but not logged.
	// MAC address (used by non-phone devices).  12 hexadecimal digits;
	// no separators (eg "0016E6513AC2", not "00:16:E6:51:3A:C2").
	MacAddr []string `protobuf:"bytes,9,rep,name=mac_addr,json=macAddr" json:"mac_addr,omitempty"` // MAC address, reported but not logged.
	// An array parallel to mac_addr, describing the type of interface.
	// Currently accepted values: "wifi", "ethernet", "bluetooth".  If
	// not present, "wifi" is assumed.
	MacAddrType []string `protobuf:"bytes,19,rep,name=mac_addr_type,json=macAddrType" json:"mac_addr_type,omitempty"`
	// Serial number (a manufacturer-defined unique hardware
	// identifier).  Alphanumeric, case-insensitive.
	SerialNumber *string `protobuf:"bytes,16,opt,name=serial_number,json=serialNumber" json:"serial_number,omitempty"`
	// Older CDMA networks use an ESN (8 hex digits) instead of an MEID.
	Esn       *string              `protobuf:"bytes,17,opt,name=esn" json:"esn,omitempty"`                              // ESN, reported but not logged
	Id        *int64               `protobuf:"varint,2,opt,name=id" json:"id,omitempty"`                                // Android device ID, not logged
	LoggingId *int64               `protobuf:"varint,7,opt,name=logging_id,json=loggingId" json:"logging_id,omitempty"` // Pseudonymous logging ID for Sawmill
	Digest    *string              `protobuf:"bytes,3,opt,name=digest" json:"digest,omitempty"`                         // Digest of device provisioning, not logged.
	Locale    *string              `protobuf:"bytes,6,opt,name=locale" json:"locale,omitempty"`                         // Current locale in standard (xx_XX) format
	Checkin   *AndroidCheckinProto `protobuf:"bytes,4,req,name=checkin" json:"checkin,omitempty"`
	// DEPRECATED, see AndroidCheckinProto.requested_group
	DesiredBuild *string `protobuf:"bytes,5,opt,name=desired_build,json=desiredBuild" json:"desired_build,omitempty"`
	// Blob of data from the Market app to be passed to Market API server
	MarketCheckin *string `protobuf:"bytes,8,opt,name=market_checkin,json=marketCheckin" json:"market_checkin,omitempty"`
	// SID cookies of any google accounts stored on the phone.  Not logged.
	AccountCookie []string `protobuf:"bytes,11,rep,name=account_cookie,json=accountCookie" json:"account_cookie,omitempty"`
	// Time zone.  Not currently logged.
	TimeZone *string `protobuf:"bytes,12,opt,name=time_zone,json=timeZone" json:"time_zone,omitempty"`
	// Security token used to validate the checkin request.
	// Required for android IDs issued to Froyo+ devices, not for legacy IDs.
	SecurityToken *uint64 `protobuf:"fixed64,13,opt,name=security_token,json=securityToken" json:"security_token,omitempty"`
	// Version of checkin protocol.
	//
	// There are currently two versions:
	//
	//   - version field missing: android IDs are assigned based on
	//     hardware identifiers.  unsecured in the sense that you can
	//     "unregister" someone's phone by sending a registration request
	//     with their IMEI/MEID/MAC.
	//
	//   - version=2: android IDs are assigned randomly.  The device is
	//     sent a security token that must be included in all future
	//     checkins for that android id.
	//
	//   - version=3: same as version 2, but the 'fragment' field is
	//     provided, and the device understands incremental updates to the
	//     gservices table (ie, only returning the keys whose values have
	//     changed.)
	//
	// (version=1 was skipped to avoid confusion with the "missing"
	// version field that is effectively version 1.)
	Version *int32 `protobuf:"varint,14,opt,name=version" json:"version,omitempty"`
	// OTA certs accepted by device (base-64 SHA-1 of cert files).  Not
	// logged.
	OtaCert []string `protobuf:"bytes,15,rep,name=ota_cert,json=otaCert" json:"ota_cert,omitempty"`
	// A single CheckinTask on the device may lead to multiple checkin
	// requests if there is too much log data to upload in a single
	// request.  For version 3 and up, this field will be filled in with
	// the number of the request, starting with 0.
	Fragment *int32 `protobuf:"varint,20,opt,name=fragment" json:"fragment,omitempty"`
	// For devices supporting multiple users, the name of the current
	// profile (they all check in independently, just as if they were
	// multiple physical devices).  This may not be set, even if the
	// device is using multiuser.  (checkin.user_number should be set to
	// the ordinal of the user.)
	UserName *string `protobuf:"bytes,21,opt,name=user_name,json=userName" json:"user_name,omitempty"`
	// For devices supporting multiple user profiles, the serial number
	// for the user checking in.  Not logged.  May not be set, even if
	// the device supportes multiuser.  checkin.user_number is the
	// ordinal of the user (0, 1, 2, ...), which may be reused if users
	// are deleted and re-created.  user_serial_number is never reused
	// (unless the device is wiped).
	UserSerialNumber *int32 `protobuf:"varint,22,opt,name=user_serial_number,json=userSerialNumber" json:"user_serial_number,omitempty"`
	unknownFields    protoimpl.UnknownFields
	sizeCache        protoimpl.SizeCache
}

func (x *AndroidCheckinRequest) Reset() {
	*x = AndroidCheckinRequest{}
	mi := &file_checkin_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AndroidCheckinRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AndroidCheckinRequest) ProtoMessage() {}

func (x *AndroidCheckinRequest) ProtoReflect() protoreflect.Message {
	mi := &file_checkin_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AndroidCheckinRequest.ProtoReflect.Descriptor instead.
func (*AndroidCheckinRequest) Descriptor() ([]byte, []int) {
	return file_checkin_proto_rawDescGZIP(), []int{1}
}

func (x *AndroidCheckinRequest) GetImei() string {
	if x != nil && x.Imei != nil {
		return *x.Imei
	}
	return ""
}

func (x *AndroidCheckinRequest) GetMeid() string {
	if x != nil && x.Meid != nil {
		return *x.Meid
	}
	return ""
}

func (x *AndroidCheckinRequest) GetMacAddr() []string {
	if x != nil {
		return x.MacAddr
	}
	return nil
}

func (x *AndroidCheckinRequest) GetMacAddrType() []string {
	if x != nil {
		return x.MacAddrType
	}
	return nil
}

func (x *AndroidCheckinRequest) GetSerialNumber() string {
	if x != nil && x.SerialNumber != nil {
		return *x.SerialNumber
	}
	return ""
}

func (x *AndroidCheckinRequest) GetEsn() string {
	if x != nil && x.Esn != nil {
		return *x.Esn
	}
	return ""
}

func (x *AndroidCheckinRequest) GetId() int64 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *AndroidCheckinRequest) GetLoggingId() int64 {
	if x != nil && x.LoggingId != nil {
		return *x.LoggingId
	}
	return 0
}

func (x *AndroidCheckinRequest) GetDigest() string {
	if x != nil && x.Digest != nil {
		return *x.Digest
	}
	return ""
}

func (x *AndroidCheckinRequest) GetLocale() string {
	if x != nil && x.Locale != nil {
		return *x.Locale
	}
	return ""
}

func (x *AndroidCheckinRequest) GetCheckin() *AndroidCheckinProto {
	if x != nil {
		return x.Checkin
	}
	return nil
}

func (x *AndroidCheckinRequest) GetDesiredBuild() string {
	if x != nil && x.DesiredBuild != nil {
		return *x.DesiredBuild
	}
	return ""
}

func (x *AndroidCheckinRequest) GetMarketCheckin() string {
	if x != nil && x.MarketCheckin != nil {
		return *x.MarketCheckin
	}
	return ""
}

func (x *AndroidCheckinRequest) GetAccountCookie() []string {
	if x != nil {
		return x.AccountCookie
	}
	return nil
}

func (x *AndroidCheckinRequest) GetTimeZone() string {
	if x != nil && x.TimeZone != nil {
		return *x.TimeZone
	}
	return ""
}

func (x *AndroidCheckinRequest) GetSecurityToken() uint64 {
	if x != nil && x.SecurityToken != nil {
		return *x.SecurityToken
	}
	return 0
}

func (x *AndroidCheckinRequest) GetVersion() int32 {
	if x != nil && x.Version != nil {
		return *x.Version
	}
	return 0
}

func (x *AndroidCheckinRequest) GetOtaCert() []string {
	if x != nil {
		return x.OtaCert
	}
	return nil
}

func (x *AndroidCheckinRequest) GetFragment() int32 {
	if x != nil && x.Fragment != nil {
		return *x.Fragment
	}
	return 0
}

func (x *AndroidCheckinRequest) GetUserName() string {
	if x != nil && x.UserName != nil {
		return *x.UserName
	}
	return ""
}

func (x *AndroidCheckinRequest) GetUserSerialNumber() int32 {
	if x != nil && x.UserSerialNumber != nil {
		return *x.UserSerialNumber
	}
	return 0
}

// The response to the device.
type AndroidCheckinResponse struct {
	state    protoimpl.MessageState `protogen:"open.v1"`
	StatsOk  *bool                  `protobuf:"varint,1,req,name=stats_ok,json=statsOk" json:"stats_ok,omitempty"`    // Whether statistics were recorded properly.
	TimeMsec *int64                 `protobuf:"varint,3,opt,name=time_msec,json=timeMsec" json:"time_msec,omitempty"` // Time of day from server (Java epoch).
	// Provisioning is sent if the request included an obsolete digest.
	//
	// For version <= 2, 'digest' contains the digest that should be
	// sent back to the server on the next checkin, and 'setting'
	// contains the entire gservices table (which replaces the entire
	// current table on the device).
	//
	// for version >= 3, 'digest' will be absent.  If 'settings_diff'
	// is false, then 'setting' contains the entire table, as in version
	// 2.  If 'settings_diff' is true, then 'delete_setting' contains
	// the keys to delete, and 'setting' contains only keys to be added
	// or for which the value has changed.  All other keys in the
	// current table should be left untouched.  If 'settings_diff' is
	// absent, don't touch the existing gservices table.
	Digest        *string             `protobuf:"bytes,4,opt,name=digest" json:"digest,omitempty"`
	SettingsDiff  *bool               `protobuf:"varint,9,opt,name=settings_diff,json=settingsDiff" json:"settings_diff,omitempty"`
	DeleteSetting []string            `protobuf:"bytes,10,rep,name=delete_setting,json=deleteSetting" json:"delete_setting,omitempty"`
	Setting       []*GservicesSetting `protobuf:"bytes,5,rep,name=setting" json:"setting,omitempty"`
	MarketOk      *bool               `protobuf:"varint,6,opt,name=market_ok,json=marketOk" json:"market_ok,omitempty"`                 // If Market got the market_checkin data OK.
	AndroidId     *uint64             `protobuf:"fixed64,7,opt,name=android_id,json=androidId" json:"android_id,omitempty"`             // From the request, or newly assigned
	SecurityToken *uint64             `protobuf:"fixed64,8,opt,name=security_token,json=securityToken" json:"security_token,omitempty"` // The associated security token
	VersionInfo   *string             `protobuf:"bytes,11,opt,name=version_info,json=versionInfo" json:"version_info,omitempty"`        // NEXT TAG: 12
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *AndroidCheckinResponse) Reset() {
	*x = AndroidCheckinResponse{}
	mi := &file_checkin_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *AndroidCheckinResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AndroidCheckinResponse) ProtoMessage() {}

func (x *AndroidCheckinResponse) ProtoReflect() protoreflect.Message {
	mi := &file_checkin_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AndroidCheckinResponse.ProtoReflect.Descriptor instead.
func (*AndroidCheckinResponse) Descriptor() ([]byte, []int) {
	return file_checkin_proto_rawDescGZIP(), []int{2}
}

func (x *AndroidCheckinResponse) GetStatsOk() bool {
	if x != nil && x.StatsOk != nil {
		return *x.StatsOk
	}
	return false
}

func (x *AndroidCheckinResponse) GetTimeMsec() int64 {
	if x != nil && x.TimeMsec != nil {
		return *x.TimeMsec
	}
	return 0
}

func (x *AndroidCheckinResponse) GetDigest() string {
	if x != nil && x.Digest != nil {
		return *x.Digest
	}
	return ""
}

func (x *AndroidCheckinResponse) GetSettingsDiff() bool {
	if x != nil && x.SettingsDiff != nil {
		return *x.SettingsDiff
	}
	return false
}

func (x *AndroidCheckinResponse) GetDeleteSetting() []string {
	if x != nil {
		return x.DeleteSetting
	}
	return nil
}

func (x *AndroidCheckinResponse) GetSetting() []*GservicesSetting {
	if x != nil {
		return x.Setting
	}
	return nil
}

func (x *AndroidCheckinResponse) GetMarketOk() bool {
	if x != nil && x.MarketOk != nil {
		return *x.MarketOk
	}
	return false
}

func (x *AndroidCheckinResponse) GetAndroidId() uint64 {
	if x != nil && x.AndroidId != nil {
		return *x.AndroidId
	}
	return 0
}

func (x *AndroidCheckinResponse) GetSecurityToken() uint64 {
	if x != nil && x.SecurityToken != nil {
		return *x.SecurityToken
	}
	return 0
}

func (x *AndroidCheckinResponse) GetVersionInfo() string {
	if x != nil && x.VersionInfo != nil {
		return *x.VersionInfo
	}
	return ""
}

var File_checkin_proto protoreflect.FileDescriptor

const file_checkin_proto_rawDesc = "" +
	"\n" +
	"\rcheckin.proto\x12\rcheckin_proto\x1a\x15android_checkin.proto\"<\n" +
	"\x10GservicesSetting\x12\x12\n" +
	"\x04name\x18\x01 \x02(\fR\x04name\x12\x14\n" +
	"\x05value\x18\x02 \x02(\fR\x05value\"\xa5\x05\n" +
	"\x15AndroidCheckinRequest\x12\x12\n" +
	"\x04imei\x18\x01 \x01(\tR\x04imei\x12\x12\n" +
	"\x04meid\x18\n" +
	" \x01(\tR\x04meid\x12\x19\n" +
	"\bmac_addr\x18\t \x03(\tR\amacAddr\x12\"\n" +
	"\rmac_addr_type\x18\x13 \x03(\tR\vmacAddrType\x12#\n" +
	"\rserial_number\x18\x10 \x01(\tR\fserialNumber\x12\x10\n" +
	"\x03esn\x18\x11 \x01(\tR\x03esn\x12\x0e\n" +
	"\x02id\x18\x02 \x01(\x03R\x02id\x12\x1d\n" +
	"\n" +
	"logging_id\x18\a \x01(\x03R\tloggingId\x12\x16\n" +
	"\x06digest\x18\x03 \x01(\tR\x06digest\x12\x16\n" +
	"\x06locale\x18\x06 \x01(\tR\x06locale\x12<\n" +
	"\acheckin\x18\x04 \x02(\v2\".checkin_proto.AndroidCheckinProtoR\acheckin\x12#\n" +
	"\rdesired_build\x18\x05 \x01(\tR\fdesiredBuild\x12%\n" +
	"\x0emarket_checkin\x18\b \x01(\tR\rmarketCheckin\x12%\n" +
	"\x0eaccount_cookie\x18\v \x03(\tR\raccountCookie\x12\x1b\n" +
	"\ttime_zone\x18\f \x01(\tR\btimeZone\x12%\n" +
	"\x0esecurity_token\x18\r \x01(\x06R\rsecurityToken\x12\x18\n" +
	"\aversion\x18\x0e \x01(\x05R\aversion\x12\x19\n" +
	"\bota_cert\x18\x0f \x03(\tR\aotaCert\x12\x1a\n" +
	"\bfragment\x18\x14 \x01(\x05R\bfragment\x12\x1b\n" +
	"\tuser_name\x18\x15 \x01(\tR\buserName\x12,\n" +
	"\x12user_serial_number\x18\x16 \x01(\x05R\x10userSerialNumber\"\xf5\x02\n" +
	"\x16AndroidCheckinResponse\x12\x19\n" +
	"\bstats_ok\x18\x01 \x02(\bR\astatsOk\x12\x1b\n" +
	"\ttime_msec\x18\x03 \x01(\x03R\btimeMsec\x12\x16\n" +
	"\x06digest\x18\x04 \x01(\tR\x06digest\x12#\n" +
	"\rsettings_diff\x18\t \x01(\bR\fsettingsDiff\x12%\n" +
	"\x0edelete_setting\x18\n" +
	" \x03(\tR\rdeleteSetting\x129\n" +
	"\asetting\x18\x05 \x03(\v2\x1f.checkin_proto.GservicesSettingR\asetting\x12\x1b\n" +
	"\tmarket_ok\x18\x06 \x01(\bR\bmarketOk\x12\x1d\n" +
	"\n" +
	"android_id\x18\a \x01(\x06R\tandroidId\x12%\n" +
	"\x0esecurity_token\x18\b \x01(\x06R\rsecurityToken\x12!\n" +
	"\fversion_info\x18\v \x01(\tR\vversionInfoB1H\x03Z-github.com/chickenfresh/go-rustplus/fcm/proto"

var (
	file_checkin_proto_rawDescOnce sync.Once
	file_checkin_proto_rawDescData []byte
)

func file_checkin_proto_rawDescGZIP() []byte {
	file_checkin_proto_rawDescOnce.Do(func() {
		file_checkin_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_checkin_proto_rawDesc), len(file_checkin_proto_rawDesc)))
	})
	return file_checkin_proto_rawDescData
}

var file_checkin_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_checkin_proto_goTypes = []any{
	(*GservicesSetting)(nil),       // 0: checkin_proto.GservicesSetting
	(*AndroidCheckinRequest)(nil),  // 1: checkin_proto.AndroidCheckinRequest
	(*AndroidCheckinResponse)(nil), // 2: checkin_proto.AndroidCheckinResponse
	(*AndroidCheckinProto)(nil),    // 3: checkin_proto.AndroidCheckinProto
}
var file_checkin_proto_depIdxs = []int32{
	3, // 0: checkin_proto.AndroidCheckinRequest.checkin:type_name -> checkin_proto.AndroidCheckinProto
	0, // 1: checkin_proto.AndroidCheckinResponse.setting:type_name -> checkin_proto.GservicesSetting
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_checkin_proto_init() }
func file_checkin_proto_init() {
	if File_checkin_proto != nil {
		return
	}
	file_android_checkin_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_checkin_proto_rawDesc), len(file_checkin_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_checkin_proto_goTypes,
		DependencyIndexes: file_checkin_proto_depIdxs,
		MessageInfos:      file_checkin_proto_msgTypes,
	}.Build()
	File_checkin_proto = out.File
	file_checkin_proto_goTypes = nil
	file_checkin_proto_depIdxs = nil
}
