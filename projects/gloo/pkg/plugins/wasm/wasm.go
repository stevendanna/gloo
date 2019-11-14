package wasm

// import (
// 	envoycore "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
// )

import (
	"github.com/gogo/protobuf/types"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

type Settings struct {
	// ref to cache cluster
	CacheRef core.ResourceRef
	// CachePath string
}

// can i fetch one of the layers?
// probably yes

type FilterCrd struct {
	Image string
}

type PluginSource struct {
	// Name of filter CRD to use.
	FilterRef core.ResourceRef
	// image name as an alternative to filter  ref
	// use filter ref for less cache misses.
	Image string `json:"image,omitempty"`

	// this should be proto any?
	// do we want to verify this if we can the proto descriptors from the image!?
	Config string `json:"config,omitempty"`

	// TODO:not a string
	FilterStage string `json:"filterStage,omitempty"`

	Name string `json:"name,omitempty"`

	// TODO:do we need this or should enforce a convention?
	// e.g. rootid == name
	RootId string `json:"root_id,omitempty"`
}

// TODO:not a string..
type Schema string

type CachedPlugin struct {
	Schema Schema
	Sha256 string
}

/*
wasmfilter init
creates a folder with all init code,
including proto buf config that one can customize
build compiles the proto and creates the required descriptors
and builds the filter with emscripten?


Image format includes:

Binary filter
Root ID
schema for config:
either:
	protobuf descriptors?
	json schema?
*/

// gloo configures the plugin. hand copy the structs until they are merged upstream and appear
// in go-control-plane

type VmConfig struct {
	// An ID which will be used along with a hash of the wasm code (or null_vm_id) to determine which
	// VM will be used for the plugin. All plugins which use the same vm_id and code will use the same
	// VM. May be left blank.
	VmID string `json:"vm_id,omitempty"`

	// The Wasm runtime type (see source/extensions/commmon/wasm/well_known_names.h).
	Runtime string `json:"runtime,omitempty"`

	// The Wasm code that Envoy will execute.
	Code *AsyncDataSource `json:"code,omitempty"`

	// The Wasm configuration string used on initialization of a new VM (proxy_onStart).
	Configuration string `json:"configuration,omitempty"`

	// Allow the wasm file to include pre-compiled code on VMs which support it.
	AllowPrecompiled bool `json:"allow_precompiled,omitempty"`
}

// Base Configuration for Wasm Plugins e.g. fiters and services.
type PluginConfig struct {
	// A unique name for a filters/services in a VM for use in identifiying the filter/service if
	// multiple filters/services are handled by the same vm_id and root_id and for logging/debugging.
	Name string `json:"name,omitempty"`

	// A unique ID for a set of filters/services in a VM which will share a RootContext and Contexts
	// if applicable (e.g. an Wasm HttpFilter and an Wasm AccessLog). If left blank, all
	// filters/services with a blank root_id with the same vm_id will share Context(s).
	RootId string `json:"root_id,omitempty"`

	// Configuration for finding or starting VM.
	VmConfig *VmConfig `json:"vm_config,omitempty"`

	// Filter/service configuration string e.g. a serialized protobuf which will be the
	// argument to the proxy_onConfigure() call.
	Configuration string `json:"configuration,omitempty"`
}

// WasmService is configured as a built-in *envoy.wasm_service* :ref:`ServiceConig
// <envoy_api_msg_config.wasm.v2.ServiceConfig>`. This opaque configuration will be used to
// create a Wasm Service.
type Wasm struct {
	// General plugin configuration.
	Config *PluginConfig `json:"config,omitempty"`
}

//////////////////////////////

// Data source consisting of either a file or an inline value.
type DataSource struct {
	// Local filesystem data source.
	Filename string `json:"filename,omitempty"`

	// Bytes inlined in the configuration.
	InlineBytes []byte `json:"inline_bytes,omitempty"`

	// String inlined in the configuration.
	InlineString string `json:"inline_string,omitempty"`
}

// This specifies how to fetch data from remote and how to verify it.
type RemoteDataSource struct {
	// The HTTP URI to fetch the remote data.
	HttpUri *HttpUri `json:"http_uri,omitempty"`

	// SHA256 string for verifying data.
	Sha256 string `json:"sha256,omitempty"`
}

// Async data source which support async data fetch.
type AsyncDataSource struct {
	// Local async data source.
	Local *DataSource `json:"local,omitempty"`

	// Remote async data source.
	Remote *RemoteDataSource `json:"remote,omitempty"`
}

type HttpUri struct {
	// The HTTP server URI. It should be a full FQDN with protocol, host and path.
	//
	// Example:
	//
	// .. code-block::yaml
	//
	//    uri:https://www.googleapis.com/oauth2/v1/certs
	//
	Uri string `json:"uri,omitempty"`

	// Specify how `uri` is to be fetched. Today, this requires an explicit
	// cluster, but in the future we may support dynamic cluster creation or
	// inline DNS resolution. See `issue
	// <https://github.com/envoyproxy/envoy/issues/1606>`_.
	// A cluster is created in the Envoy "cluster_manager" config
	// section. This field specifies the cluster name.
	//
	// Example:
	//
	// .. code-block::yaml
	//
	//    cluster:jwks_cluster
	//
	Cluster string `json:"cluster,omitempty"`

	// Sets the maximum duration in milliseconds that a response can take to arrive upon request.
	Timeout *types.Duration `json:"timeout,omitempty"`
}
